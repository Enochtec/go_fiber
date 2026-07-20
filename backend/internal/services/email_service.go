package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// ─── Types ────────────────────────────────────────────────────────────────────

type ReceiptEmailData struct {
	ToEmail      string
	ToName       string
	ShopName     string
	ShopAddress  string
	ShopPhone    string
	ShopEmail    string
	ShopLogo     string
	ReceiptNo    string
	Date         string
	Time         string
	Cashier      string
	Customer     string
	PaymentMethod string
	Items        []ReceiptEmailItem
	Subtotal     float64
	Discount     float64
	Tax          float64
	Total        float64
	Tendered     float64
	Change       float64
	Footer       string
	Currency     string
}

type ReceiptEmailItem struct {
	Name      string
	Quantity  int
	UnitPrice float64
	Total     float64
}

type EmailService struct{}

func NewEmailService() *EmailService { return &EmailService{} }

// ─── Send helpers ─────────────────────────────────────────────────────────────

func (s *EmailService) send(to, subject, html string) error {
	apiKey := os.Getenv("RESEND_API_KEY")
	if apiKey == "" {
		return fmt.Errorf("RESEND_API_KEY is not set")
	}
	fromAddr := os.Getenv("RESEND_FROM_EMAIL")
	if fromAddr == "" {
		fromAddr = "noreply@maestropos.app"
	}
	fromName := os.Getenv("RESEND_FROM_NAME")
	if fromName == "" {
		fromName = "Maestro POS"
	}

	payload, _ := json.Marshal(map[string]interface{}{
		"from":    fromName + " <" + fromAddr + ">",
		"to":      []string{to},
		"subject": subject,
		"html":    html,
	})

	req, _ := http.NewRequest("POST", "https://api.resend.com/emails", bytes.NewReader(payload))
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("resend request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("resend error %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

// ─── Receipt email ────────────────────────────────────────────────────────────

func (s *EmailService) SendReceipt(data ReceiptEmailData) error {
	subject := fmt.Sprintf("Your Receipt #%s — %s", data.ReceiptNo, data.ShopName)
	html    := buildReceiptHTML(data)
	return s.send(data.ToEmail, subject, html)
}

// ─── HTML template ────────────────────────────────────────────────────────────

func fmtMoney(n float64, currency string) string {
	if currency == "" {
		currency = "KES"
	}
	return fmt.Sprintf("%s %.2f", currency, n)
}

func buildReceiptHTML(d ReceiptEmailData) string {
	if d.Currency == "" {
		d.Currency = "KES"
	}

	// Build items rows
	var itemRows strings.Builder
	for _, item := range d.Items {
		itemRows.WriteString(fmt.Sprintf(`
			<tr>
				<td style="padding:8px 12px;border-bottom:1px solid #f1f5f9;font-size:13px;color:#1e293b;">%s</td>
				<td style="padding:8px 12px;border-bottom:1px solid #f1f5f9;font-size:13px;color:#64748b;text-align:center;">%d</td>
				<td style="padding:8px 12px;border-bottom:1px solid #f1f5f9;font-size:13px;color:#64748b;text-align:right;">%.2f</td>
				<td style="padding:8px 12px;border-bottom:1px solid #f1f5f9;font-size:13px;font-weight:600;color:#1e293b;text-align:right;">%.2f</td>
			</tr>`,
			escapeHTML(item.Name), item.Quantity, item.UnitPrice, item.Total))
	}

	// Discount/Tax rows (conditional)
	var discountRow, taxRow string
	if d.Discount > 0 {
		discountRow = fmt.Sprintf(`
		<tr>
			<td colspan="3" style="padding:4px 12px;font-size:12px;color:#64748b;text-align:right;">Discount</td>
			<td style="padding:4px 12px;font-size:12px;color:#16a34a;text-align:right;">-%s</td>
		</tr>`, fmtMoney(d.Discount, d.Currency))
	}
	if d.Tax > 0 {
		taxRow = fmt.Sprintf(`
		<tr>
			<td colspan="3" style="padding:4px 12px;font-size:12px;color:#64748b;text-align:right;">Tax</td>
			<td style="padding:4px 12px;font-size:12px;color:#64748b;text-align:right;">%s</td>
		</tr>`, fmtMoney(d.Tax, d.Currency))
	}

	// Cash tendered rows
	var cashRows string
	if strings.ToLower(d.PaymentMethod) == "cash" && d.Tendered > 0 {
		cashRows = fmt.Sprintf(`
		<tr style="background:#f8fafc;">
			<td colspan="3" style="padding:6px 12px;font-size:12px;color:#64748b;text-align:right;">Tendered</td>
			<td style="padding:6px 12px;font-size:12px;color:#64748b;text-align:right;">%s</td>
		</tr>
		<tr style="background:#f8fafc;">
			<td colspan="3" style="padding:6px 12px;font-size:12px;color:#64748b;text-align:right;">Change</td>
			<td style="padding:6px 12px;font-size:12px;color:#64748b;text-align:right;">%s</td>
		</tr>`, fmtMoney(d.Tendered, d.Currency), fmtMoney(d.Change, d.Currency))
	}

	// Logo section
	logoSection := ""
	if d.ShopLogo != "" {
		logoSection = fmt.Sprintf(`<img src="%s" alt="%s" style="height:48px;width:auto;object-fit:contain;margin-bottom:8px;" />`,
			escapeHTML(d.ShopLogo), escapeHTML(d.ShopName))
	}

	// Contact info
	var contactParts []string
	if d.ShopPhone != "" {
		contactParts = append(contactParts, escapeHTML(d.ShopPhone))
	}
	if d.ShopEmail != "" {
		contactParts = append(contactParts, escapeHTML(d.ShopEmail))
	}
	contactLine := strings.Join(contactParts, " &nbsp;|&nbsp; ")

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head><meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>Receipt</title></head>
<body style="margin:0;padding:0;background:#f1f5f9;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif;">
<table width="100%%" cellpadding="0" cellspacing="0" style="background:#f1f5f9;padding:32px 0;">
<tr><td align="center">
<table width="560" cellpadding="0" cellspacing="0" style="background:#ffffff;border-radius:4px;overflow:hidden;box-shadow:0 1px 3px rgba(0,0,0,.1);">

  <!-- Header -->
  <tr>
    <td style="background:#0f172a;padding:28px 32px;text-align:center;">
      %s
      <h1 style="margin:0;color:#ffffff;font-size:20px;font-weight:700;letter-spacing:.5px;">%s</h1>
      %s
      %s
    </td>
  </tr>

  <!-- Receipt tag -->
  <tr>
    <td style="background:#008B8B;padding:8px 32px;text-align:center;">
      <p style="margin:0;color:#ffffff;font-size:11px;font-weight:600;letter-spacing:2px;text-transform:uppercase;">OFFICIAL RECEIPT</p>
    </td>
  </tr>

  <!-- Meta info -->
  <tr>
    <td style="padding:20px 32px 0;">
      <table width="100%%" cellpadding="0" cellspacing="0">
        <tr>
          <td width="50%%">
            <p style="margin:0 0 4px;font-size:11px;color:#94a3b8;text-transform:uppercase;letter-spacing:.5px;">Receipt No.</p>
            <p style="margin:0;font-size:14px;font-weight:700;color:#1e293b;font-family:monospace;">#%s</p>
          </td>
          <td width="50%%" align="right">
            <p style="margin:0 0 4px;font-size:11px;color:#94a3b8;text-transform:uppercase;letter-spacing:.5px;">Date &amp; Time</p>
            <p style="margin:0;font-size:14px;color:#1e293b;">%s &nbsp; %s</p>
          </td>
        </tr>
        <tr><td colspan="2" style="padding-top:12px;"></td></tr>
        <tr>
          <td width="50%%">
            <p style="margin:0 0 4px;font-size:11px;color:#94a3b8;text-transform:uppercase;letter-spacing:.5px;">Cashier</p>
            <p style="margin:0;font-size:13px;color:#1e293b;">%s</p>
          </td>
          <td width="50%%" align="right">
            <p style="margin:0 0 4px;font-size:11px;color:#94a3b8;text-transform:uppercase;letter-spacing:.5px;">Customer</p>
            <p style="margin:0;font-size:13px;color:#1e293b;">%s</p>
          </td>
        </tr>
        <tr><td colspan="2" style="padding-top:12px;"></td></tr>
        <tr>
          <td colspan="2">
            <p style="margin:0 0 4px;font-size:11px;color:#94a3b8;text-transform:uppercase;letter-spacing:.5px;">Payment Method</p>
            <p style="margin:0;font-size:13px;font-weight:600;color:#008B8B;text-transform:uppercase;">%s</p>
          </td>
        </tr>
      </table>
    </td>
  </tr>

  <!-- Divider -->
  <tr><td style="padding:16px 32px;"><hr style="border:0;border-top:1px solid #e2e8f0;"></td></tr>

  <!-- Items -->
  <tr>
    <td style="padding:0 32px;">
      <table width="100%%" cellpadding="0" cellspacing="0" style="border-collapse:collapse;">
        <thead>
          <tr style="background:#f8fafc;">
            <th style="padding:8px 12px;font-size:11px;color:#94a3b8;text-align:left;font-weight:600;text-transform:uppercase;letter-spacing:.5px;">Item</th>
            <th style="padding:8px 12px;font-size:11px;color:#94a3b8;text-align:center;font-weight:600;text-transform:uppercase;letter-spacing:.5px;">Qty</th>
            <th style="padding:8px 12px;font-size:11px;color:#94a3b8;text-align:right;font-weight:600;text-transform:uppercase;letter-spacing:.5px;">Price</th>
            <th style="padding:8px 12px;font-size:11px;color:#94a3b8;text-align:right;font-weight:600;text-transform:uppercase;letter-spacing:.5px;">Total</th>
          </tr>
        </thead>
        <tbody>%s</tbody>
        <tfoot>
          <tr>
            <td colspan="3" style="padding:8px 12px;font-size:12px;color:#64748b;text-align:right;">Subtotal</td>
            <td style="padding:8px 12px;font-size:12px;color:#1e293b;text-align:right;">%s</td>
          </tr>
          %s
          %s
          <tr style="background:#0f172a;">
            <td colspan="3" style="padding:12px;font-size:14px;font-weight:700;color:#ffffff;text-align:right;">TOTAL</td>
            <td style="padding:12px;font-size:16px;font-weight:700;color:#ffffff;text-align:right;">%s</td>
          </tr>
          %s
        </tfoot>
      </table>
    </td>
  </tr>

  <!-- Footer -->
  <tr>
    <td style="padding:24px 32px;text-align:center;border-top:1px solid #e2e8f0;margin-top:16px;">
      <p style="margin:0 0 4px;font-size:13px;color:#64748b;">%s</p>
      <p style="margin:0;font-size:11px;color:#94a3b8;">Powered by Maestro POS</p>
    </td>
  </tr>

</table>
</td></tr>
</table>
</body>
</html>`,
		logoSection,
		escapeHTML(d.ShopName),
		func() string {
			if d.ShopAddress != "" {
				return `<p style="margin:4px 0 0;color:#94a3b8;font-size:12px;">` + escapeHTML(d.ShopAddress) + `</p>`
			}
			return ""
		}(),
		func() string {
			if contactLine != "" {
				return `<p style="margin:4px 0 0;color:#94a3b8;font-size:12px;">` + contactLine + `</p>`
			}
			return ""
		}(),
		escapeHTML(d.ReceiptNo),
		escapeHTML(d.Date), escapeHTML(d.Time),
		escapeHTML(d.Cashier),
		func() string {
			if d.Customer != "" {
				return escapeHTML(d.Customer)
			}
			return "Walk-in"
		}(),
		escapeHTML(strings.ToUpper(d.PaymentMethod)),
		itemRows.String(),
		fmtMoney(d.Subtotal, d.Currency),
		discountRow, taxRow,
		fmtMoney(d.Total, d.Currency),
		cashRows,
		func() string {
			if d.Footer != "" {
				return escapeHTML(d.Footer)
			}
			return "Thank you for your business!"
		}(),
	)
}

func escapeHTML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, `"`, "&#34;")
	return s
}
