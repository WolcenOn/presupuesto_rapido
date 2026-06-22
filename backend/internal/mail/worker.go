package mail

import (
	"context"
	"fmt"
	"log"
	"net/smtp"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Config struct {
	Host      string
	Port      int
	Username  string
	Password  string
	FromEmail string
	FromName  string
}

type Worker struct {
	DB     *pgxpool.Pool
	Config Config
}

func (w Worker) Start(ctx context.Context) {
	if w.DB == nil || strings.TrimSpace(w.Config.Host) == "" || strings.TrimSpace(w.Config.FromEmail) == "" {
		log.Println("mail worker disabled: database, SMTP host or from address missing")
		return
	}
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			w.processOnce(ctx)
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
			}
		}
	}()
}

func (w Worker) processOnce(ctx context.Context) {
	rows, err := w.DB.Query(ctx, `
		select l.id, l.recipient, d.ref, d.type, d.client_name, d.total_amount
		from document_email_logs l
		join documents d on d.id = l.document_id
		where l.status = 'queued'
		order by l.created_at asc
		limit 10`)
	if err != nil {
		log.Printf("mail worker query failed: %v", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var id int64
		var recipient, ref, docType, clientName string
		var total float64
		if err := rows.Scan(&id, &recipient, &ref, &docType, &clientName, &total); err != nil {
			log.Printf("mail worker scan failed: %v", err)
			continue
		}
		if err := w.send(recipient, ref, docType, clientName, total); err != nil {
			_, _ = w.DB.Exec(ctx, `update document_email_logs set status = 'failed', error_message = $1 where id = $2`, err.Error(), id)
			continue
		}
		_, _ = w.DB.Exec(ctx, `
			update document_email_logs set status = 'sent', error_message = null where id = $1;
			update documents set sent_to_boss_at = now() where id = (select document_id from document_email_logs where id = $1)`, id)
	}
}

func (w Worker) send(to, ref, docType, clientName string, total float64) error {
	addr := fmt.Sprintf("%s:%d", w.Config.Host, w.Config.Port)
	subject := fmt.Sprintf("Nuevo %s %s", docType, ref)
	body := fmt.Sprintf("Se ha creado un nuevo documento.\n\nTipo: %s\nReferencia: %s\nCliente: %s\nTotal: %.2f\n", docType, ref, clientName, total)
	from := w.Config.FromEmail
	if strings.TrimSpace(w.Config.FromName) != "" {
		from = fmt.Sprintf("%s <%s>", w.Config.FromName, w.Config.FromEmail)
	}
	message := strings.Join([]string{
		"From: " + from,
		"To: " + to,
		"Subject: " + subject,
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=utf-8",
		"",
		body,
	}, "\r\n")
	var auth smtp.Auth
	if strings.TrimSpace(w.Config.Username) != "" {
		auth = smtp.PlainAuth("", w.Config.Username, w.Config.Password, w.Config.Host)
	}
	return smtp.SendMail(addr, auth, w.Config.FromEmail, []string{to}, []byte(message))
}
