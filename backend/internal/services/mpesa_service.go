package services

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"
)

// ─── Token cache ────────────────────────────────────────────────────────────

type mpesaToken struct {
	value   string
	expires time.Time
	mu      sync.Mutex
}

var mpesaTokenCache = &mpesaToken{}

// ─── Public types ────────────────────────────────────────────────────────────

type STKPushInput struct {
	ShopID    string
	Phone     string
	Amount    float64
	Reference string
	Desc      string
}

type STKPushResult struct {
	MerchantRequestID  string `json:"merchant_request_id"`
	CheckoutRequestID  string `json:"checkout_request_id"`
	ResponseCode       string `json:"response_code"`
	CustomerMessage    string `json:"customer_message"`
}

type MpesaCallbackBody struct {
	Body struct {
		StkCallback struct {
			MerchantRequestID string `json:"MerchantRequestID"`
			CheckoutRequestID string `json:"CheckoutRequestID"`
			ResultCode        int    `json:"ResultCode"`
			ResultDesc        string `json:"ResultDesc"`
			CallbackMetadata  *struct {
				Item []struct {
					Name  string      `json:"Name"`
					Value interface{} `json:"Value"`
				} `json:"Item"`
			} `json:"CallbackMetadata"`
		} `json:"stkCallback"`
	} `json:"Body"`
}

type MpesaService struct{}

func NewMpesaService() *MpesaService { return &MpesaService{} }

// ─── Helpers ─────────────────────────────────────────────────────────────────

func (s *MpesaService) baseURL() string {
	if os.Getenv("MPESA_ENV") == "production" {
		return "https://api.safaricom.co.ke"
	}
	return "https://sandbox.safaricom.co.ke"
}

func (s *MpesaService) shortcode() string {
	if v := os.Getenv("MPESA_SHORTCODE"); v != "" {
		return v
	}
	return "174379" // default sandbox
}

func (s *MpesaService) passkey() string {
	if v := os.Getenv("MPESA_PASSKEY"); v != "" {
		return v
	}
	// Safaricom sandbox passkey
	return "bfb279f9aa9bdbcf158e97dd71a467cd2e0c893059b10f78e6b72ada1ed2c919"
}

func (s *MpesaService) callbackURL() string {
	return os.Getenv("MPESA_CALLBACK_URL")
}

func (s *MpesaService) timestamp() string {
	return time.Now().Format("20060102150405")
}

func (s *MpesaService) password(ts string) string {
	raw := s.shortcode() + s.passkey() + ts
	return base64.StdEncoding.EncodeToString([]byte(raw))
}

// normalizePhone converts 07xx → 2547xx, +2547xx → 2547xx.
func normalizePhone(phone string) string {
	p := ""
	for _, c := range phone {
		if c >= '0' && c <= '9' {
			p += string(c)
		}
	}
	if len(p) == 9 { // 7xxxxxxxx
		return "254" + p
	}
	if len(p) == 10 && p[0] == '0' { // 07xxxxxxxx
		return "254" + p[1:]
	}
	if len(p) == 12 && p[:3] == "254" {
		return p
	}
	return p
}

// ─── Auth Token ──────────────────────────────────────────────────────────────

func (s *MpesaService) getAccessToken() (string, error) {
	mpesaTokenCache.mu.Lock()
	defer mpesaTokenCache.mu.Unlock()

	if mpesaTokenCache.value != "" && time.Now().Before(mpesaTokenCache.expires) {
		return mpesaTokenCache.value, nil
	}

	key    := os.Getenv("MPESA_CONSUMER_KEY")
	secret := os.Getenv("MPESA_CONSUMER_SECRET")
	if key == "" || secret == "" {
		return "", fmt.Errorf("MPESA_CONSUMER_KEY and MPESA_CONSUMER_SECRET must be set")
	}

	url := s.baseURL() + "/oauth/v1/generate?grant_type=client_credentials"
	req, _ := http.NewRequest("GET", url, nil)
	req.SetBasicAuth(key, secret)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("mpesa auth request failed: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   string `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("mpesa auth decode failed: %w", err)
	}
	if result.AccessToken == "" {
		return "", fmt.Errorf("mpesa auth: empty access token in response")
	}

	mpesaTokenCache.value   = result.AccessToken
	mpesaTokenCache.expires = time.Now().Add(55 * time.Minute)
	return result.AccessToken, nil
}

// ─── STK Push ────────────────────────────────────────────────────────────────

func (s *MpesaService) InitiateSTKPush(input STKPushInput) (*STKPushResult, error) {
	token, err := s.getAccessToken()
	if err != nil {
		return nil, err
	}

	ts    := s.timestamp()
	phone := normalizePhone(input.Phone)

	ref := input.Reference
	if ref == "" {
		ref = "POS-" + time.Now().Format("20060102150405")
	}
	desc := input.Desc
	if desc == "" {
		desc = "Payment for goods"
	}
	if len(desc) > 13 {
		desc = desc[:13]
	}

	amount := int(input.Amount)
	if amount < 1 {
		amount = 1
	}

	body := map[string]interface{}{
		"BusinessShortCode": s.shortcode(),
		"Password":          s.password(ts),
		"Timestamp":         ts,
		"TransactionType":   "CustomerPayBillOnline",
		"Amount":            amount,
		"PartyA":            phone,
		"PartyB":            s.shortcode(),
		"PhoneNumber":       phone,
		"CallBackURL":       s.callbackURL(),
		"AccountReference":  ref,
		"TransactionDesc":   desc,
	}

	payload, _ := json.Marshal(body)
	url         := s.baseURL() + "/mpesa/stkpush/v1/processrequest"
	req, _      := http.NewRequest("POST", url, bytes.NewReader(payload))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("stk push request failed: %w", err)
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)

	var result struct {
		MerchantRequestID  string `json:"MerchantRequestID"`
		CheckoutRequestID  string `json:"CheckoutRequestID"`
		ResponseCode       string `json:"ResponseCode"`
		ResponseDescription string `json:"ResponseDescription"`
		CustomerMessage    string `json:"CustomerMessage"`
		ErrorCode          string `json:"errorCode"`
		ErrorMessage       string `json:"errorMessage"`
		RequestID          string `json:"requestId"`
	}
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("stk push decode failed: %w", err)
	}
	if result.ResponseCode != "0" {
		msg := result.ErrorMessage
		if msg == "" {
			msg = result.ResponseDescription
		}
		return nil, fmt.Errorf("stk push rejected: %s", msg)
	}

	return &STKPushResult{
		MerchantRequestID: result.MerchantRequestID,
		CheckoutRequestID: result.CheckoutRequestID,
		ResponseCode:      result.ResponseCode,
		CustomerMessage:   result.CustomerMessage,
	}, nil
}

// ─── Parse Callback ───────────────────────────────────────────────────────────

type CallbackResult struct {
	CheckoutRequestID string
	MerchantRequestID string
	ResultCode        int
	ResultDesc        string
	MpesaReceipt      string
	Amount            float64
	Phone             string
}

func (s *MpesaService) ParseCallback(body MpesaCallbackBody) CallbackResult {
	cb := body.Body.StkCallback
	res := CallbackResult{
		CheckoutRequestID: cb.CheckoutRequestID,
		MerchantRequestID: cb.MerchantRequestID,
		ResultCode:        cb.ResultCode,
		ResultDesc:        cb.ResultDesc,
	}
	if cb.CallbackMetadata != nil {
		for _, item := range cb.CallbackMetadata.Item {
			switch item.Name {
			case "MpesaReceiptNumber":
				if v, ok := item.Value.(string); ok {
					res.MpesaReceipt = v
				}
			case "Amount":
				switch v := item.Value.(type) {
				case float64:
					res.Amount = v
				}
			case "PhoneNumber":
				switch v := item.Value.(type) {
				case float64:
					res.Phone = fmt.Sprintf("%.0f", v)
				case string:
					res.Phone = v
				}
			}
		}
	}
	return res
}
