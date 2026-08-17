package ups

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/hiscaler/ups-go/config"
	"gopkg.in/guregu/null.v4"
)

func TestAuthToken(t *testing.T) {
	if client == nil {
		t.Skip("client not initialized")
	}
	token, err := client.Services.Auth.Token(ctx)
	if err != nil {
		t.Fatalf("Token error: %v", err)
	}
	if token.AccessToken == "" {
		t.Fatal("access_token is empty")
	}
	t.Logf("token_type=%s expires_in=%s", token.TokenType, token.ExpiresIn)
}

func TestCreateShipmentValidation(t *testing.T) {
	req := CreateShipmentRequest{}
	err := req.Validate()
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestCreateShipmentRequestValidateOK(t *testing.T) {
	if err := validCreateShipmentRequest().Validate(); err != nil {
		t.Fatalf("unexpected validation error: %v", err)
	}
}

func TestCancelValidation(t *testing.T) {
	req := CancelShipmentRequest{}
	if err := req.Validate(); err == nil {
		t.Fatal("expected validation error")
	}
}

func TestShippingLabelValidation(t *testing.T) {
	req := ShippingLabelRequest{
		LabelSpecification: LabelSpecification{
			LabelImageFormat: LabelImageFormat{Code: "GIF"},
		},
	}
	if err := req.Validate(); err == nil {
		t.Fatal("expected validation error for empty tracking number")
	}
}

func TestShipmentCreate_InvalidRequest(t *testing.T) {
	svc := newTestShipmentService(t, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("HTTP should not be called for invalid request")
	}))

	_, err := svc.Create(ctx, CreateShipmentRequest{})
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestShipmentCreate_Success(t *testing.T) {
	var cap httpCapture
	svc := newTestShipmentService(t, captureHandler(&cap, http.StatusOK, `{
		"ShipmentResponse": {
			"Response": {"ResponseStatus": {"Code": "1", "Description": "Success"}},
			"ShipmentResults": {
				"ShipmentIdentificationNumber": "1Z9999999999999999",
				"BillingWeight": {
					"UnitOfMeasurement": {"Code": "LBS", "Description": "Pounds"},
					"Weight": "10.0"
				},
				"ShipmentCharges": {
					"TransportationCharges": {"CurrencyCode": "USD", "MonetaryValue": "12.34"},
					"TotalCharges": {"CurrencyCode": "USD", "MonetaryValue": "12.34"}
				},
				"PackageResults": {
					"TrackingNumber": "1Z9999999999999999",
					"ShippingLabel": {
						"ImageFormat": {"Code": "GIF", "Description": "GIF"},
						"GraphicImage": "R0lGODlh"
					}
				}
			}
		}
	}`))

	got, err := svc.Create(ctx, validCreateShipmentRequest())
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}

	if cap.method != http.MethodPost {
		t.Fatalf("method = %s, want POST", cap.method)
	}
	if cap.path != "/api/shipments/v2409/ship" {
		t.Fatalf("path = %s, want /api/shipments/v2409/ship", cap.path)
	}
	if cap.query.Get("additionaladdressvalidation") != "city" {
		t.Fatalf("query additionaladdressvalidation = %q", cap.query.Get("additionaladdressvalidation"))
	}
	if auth := cap.header.Get("Authorization"); auth != "Bearer test-token" {
		t.Fatalf("Authorization = %q, want Bearer test-token", auth)
	}
	if src := cap.header.Get("transactionSrc"); src != "ups-go-test" {
		t.Fatalf("transactionSrc = %q, want ups-go-test", src)
	}

	var wrapped shipmentRequestWrapper
	if err = json.Unmarshal(cap.body, &wrapped); err != nil {
		t.Fatalf("unmarshal request body: %v", err)
	}
	if wrapped.ShipmentRequest.Shipment.Shipper.Name != "Shipper Name" {
		t.Fatalf("request shipper name = %q", wrapped.ShipmentRequest.Shipment.Shipper.Name)
	}

	if got.ShipmentIdentificationNumber != "1Z9999999999999999" {
		t.Fatalf("ShipmentIdentificationNumber = %q", got.ShipmentIdentificationNumber)
	}
	if got.BillingWeight == nil || got.BillingWeight.Weight != "10.0" {
		t.Fatalf("BillingWeight = %+v", got.BillingWeight)
	}
	if got.ShipmentCharges == nil || got.ShipmentCharges.TotalCharges == nil || got.ShipmentCharges.TotalCharges.MonetaryValue != "12.34" {
		t.Fatalf("ShipmentCharges = %+v", got.ShipmentCharges)
	}
	if len(got.PackageResults) != 1 {
		t.Fatalf("PackageResults len = %d, want 1", len(got.PackageResults))
	}
	pkg := got.PackageResults[0]
	if pkg.TrackingNumber != "1Z9999999999999999" {
		t.Fatalf("TrackingNumber = %q", pkg.TrackingNumber)
	}
	if pkg.ShippingLabel == nil || pkg.ShippingLabel.GraphicImage != "R0lGODlh" {
		t.Fatalf("ShippingLabel = %+v", pkg.ShippingLabel)
	}
}

func TestShipmentCreate_PackageResultsArray(t *testing.T) {
	svc := newTestShipmentService(t, staticJSONHandler(http.StatusOK, `{
		"ShipmentResponse": {
			"Response": {"ResponseStatus": {"Code": "1", "Description": "Success"}},
			"ShipmentResults": {
				"ShipmentIdentificationNumber": "1ZSHIPMENT",
				"PackageResults": [
					{"TrackingNumber": "1ZPKG001"},
					{"TrackingNumber": "1ZPKG002"}
				]
			}
		}
	}`))

	got, err := svc.Create(ctx, validCreateShipmentRequest())
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}
	if len(got.PackageResults) != 2 {
		t.Fatalf("PackageResults len = %d, want 2", len(got.PackageResults))
	}
	if got.PackageResults[0].TrackingNumber != "1ZPKG001" || got.PackageResults[1].TrackingNumber != "1ZPKG002" {
		t.Fatalf("PackageResults = %+v", got.PackageResults)
	}
}

func TestShipmentCreate_NonSuccessStatus(t *testing.T) {
	svc := newTestShipmentService(t, staticJSONHandler(http.StatusOK, `{
		"ShipmentResponse": {
			"Response": {"ResponseStatus": {"Code": "0", "Description": "Failure"}}
		}
	}`))

	_, err := svc.Create(ctx, validCreateShipmentRequest())
	if err == nil {
		t.Fatal("expected status error")
	}
	if !strings.Contains(err.Error(), "0 Failure") {
		t.Fatalf("error = %q, want contain %q", err.Error(), "0 Failure")
	}
}

func TestShipmentCreate_UPSError(t *testing.T) {
	svc := newTestShipmentService(t, staticJSONHandler(http.StatusBadRequest, `{
		"response": {"errors": [{"code": "120001", "message": "Missing or invalid shipper number"}]}
	}`))

	_, err := svc.Create(ctx, validCreateShipmentRequest())
	if err == nil {
		t.Fatal("expected UPS error")
	}
	if !strings.Contains(err.Error(), "120001") {
		t.Fatalf("error = %q, want contain 120001", err.Error())
	}
}

func TestShipmentCancel_InvalidRequest(t *testing.T) {
	svc := newTestShipmentService(t, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("HTTP should not be called for invalid request")
	}))

	_, err := svc.Cancel(ctx, CancelShipmentRequest{})
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestShipmentCancel_Success(t *testing.T) {
	var cap httpCapture
	svc := newTestShipmentService(t, captureHandler(&cap, http.StatusOK, `{
		"VoidShipmentResponse": {
			"Response": {"ResponseStatus": {"Code": "1", "Description": "Success"}},
			"SummaryResult": {"Status": {"Code": "1", "Description": "Voided"}},
			"PackageLevelResult": {
				"TrackingNumber": "1ZPKG001",
				"Status": {"Code": "1", "Description": "Voided"}
			}
		}
	}`))

	got, err := svc.Cancel(ctx, CancelShipmentRequest{ShipmentIdentificationNumber: "1zshipmentid"})
	if err != nil {
		t.Fatalf("Cancel error: %v", err)
	}

	if cap.method != http.MethodDelete {
		t.Fatalf("method = %s, want DELETE", cap.method)
	}
	if cap.path != "/api/shipments/v2409/void/cancel/1ZSHIPMENTID" {
		t.Fatalf("path = %s, want uppercased shipment id", cap.path)
	}
	if cap.query.Get("trackingnumber") != "" {
		t.Fatalf("trackingnumber query = %q, want empty", cap.query.Get("trackingnumber"))
	}
	if got.SummaryStatusCode != "1" || got.SummaryStatusDescription != "Voided" {
		t.Fatalf("summary = %+v", got)
	}
	if len(got.PackageLevelResults) != 1 || got.PackageLevelResults[0].TrackingNumber != "1ZPKG001" {
		t.Fatalf("PackageLevelResults = %+v", got.PackageLevelResults)
	}
}

func TestShipmentCancel_TrackingNumbers(t *testing.T) {
	tests := []struct {
		name      string
		tracking  []string
		wantQuery string
		response  string
		wantPkgs  int
	}{
		{
			name:      "single",
			tracking:  []string{"1zpkg001"},
			wantQuery: "1ZPKG001",
			response: `{
				"VoidShipmentResponse": {
					"Response": {"ResponseStatus": {"Code": "1", "Description": "Success"}},
					"SummaryResult": {"Status": {"Code": "1", "Description": "Voided"}},
					"PackageLevelResult": {"TrackingNumber": "1ZPKG001", "Status": {"Code": "1", "Description": "Voided"}}
				}
			}`,
			wantPkgs: 1,
		},
		{
			name:      "multiple",
			tracking:  []string{"1zpkg001", "1zpkg002"},
			wantQuery: `["1ZPKG001","1ZPKG002"]`,
			response: `{
				"VoidShipmentResponse": {
					"Response": {"ResponseStatus": {"Code": "1", "Description": "Success"}},
					"SummaryResult": {"Status": {"Code": "1", "Description": "Voided"}},
					"PackageLevelResult": [
						{"TrackingNumber": "1ZPKG001", "Status": {"Code": "1", "Description": "Voided"}},
						{"TrackingNumber": "1ZPKG002", "Status": {"Code": "1", "Description": "Voided"}}
					]
				}
			}`,
			wantPkgs: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cap httpCapture
			svc := newTestShipmentService(t, captureHandler(&cap, http.StatusOK, tt.response))
			got, err := svc.Cancel(ctx, CancelShipmentRequest{
				ShipmentIdentificationNumber: "1ZSHIPMENT",
				TrackingNumbers:              tt.tracking,
			})
			if err != nil {
				t.Fatalf("Cancel error: %v", err)
			}
			if cap.query.Get("trackingnumber") != tt.wantQuery {
				t.Fatalf("trackingnumber = %q, want %q", cap.query.Get("trackingnumber"), tt.wantQuery)
			}
			if len(got.PackageLevelResults) != tt.wantPkgs {
				t.Fatalf("PackageLevelResults len = %d, want %d", len(got.PackageLevelResults), tt.wantPkgs)
			}
		})
	}
}

func TestShipmentCancel_NonSuccessStatus(t *testing.T) {
	svc := newTestShipmentService(t, staticJSONHandler(http.StatusOK, `{
		"VoidShipmentResponse": {
			"Response": {"ResponseStatus": {"Code": "0", "Description": "Unable to void"}}
		}
	}`))

	_, err := svc.Cancel(ctx, CancelShipmentRequest{ShipmentIdentificationNumber: "1ZSHIPMENT"})
	if err == nil {
		t.Fatal("expected status error")
	}
	if !strings.Contains(err.Error(), "Unable to void") {
		t.Fatalf("error = %q", err.Error())
	}
}

func TestShipmentShippingLabel_InvalidRequest(t *testing.T) {
	svc := newTestShipmentService(t, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("HTTP should not be called for invalid request")
	}))

	_, err := svc.ShippingLabel(ctx, ShippingLabelRequest{
		LabelSpecification: LabelSpecification{LabelImageFormat: LabelImageFormat{Code: "GIF"}},
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestShipmentShippingLabel_Success(t *testing.T) {
	var cap httpCapture
	svc := newTestShipmentService(t, captureHandler(&cap, http.StatusOK, `{
		"LabelRecoveryResponse": {
			"Response": {"ResponseStatus": {"Code": "1", "Description": "Success"}},
			"ShipmentIdentificationNumber": "1ZSHIPMENT",
			"LabelResults": {
				"TrackingNumber": "1ZPKG001",
				"LabelImage": {
					"LabelImageFormat": {"Code": "GIF", "Description": "GIF"},
					"GraphicImage": "R0lGODlh",
					"HTMLImage": "<html/>"
				}
			}
		}
	}`))

	got, err := svc.ShippingLabel(ctx, ShippingLabelRequest{
		TrackingNumber: "1zpkg001",
		LabelSpecification: LabelSpecification{
			LabelImageFormat: LabelImageFormat{Code: "GIF"},
		},
		SubVersion: null.StringFrom("1801"),
	})
	if err != nil {
		t.Fatalf("ShippingLabel error: %v", err)
	}

	if cap.method != http.MethodPost {
		t.Fatalf("method = %s, want POST", cap.method)
	}
	if cap.path != "/api/labels/v1/recovery" {
		t.Fatalf("path = %s, want /api/labels/v1/recovery", cap.path)
	}

	var wrapped labelRecoveryRequestWrapper
	if err = json.Unmarshal(cap.body, &wrapped); err != nil {
		t.Fatalf("unmarshal request body: %v", err)
	}
	if wrapped.LabelRecoveryRequest.TrackingNumber != "1ZPKG001" {
		t.Fatalf("request TrackingNumber = %q, want 1ZPKG001", wrapped.LabelRecoveryRequest.TrackingNumber)
	}
	if wrapped.LabelRecoveryRequest.Request == nil || wrapped.LabelRecoveryRequest.Request.SubVersion.String != "1801" {
		t.Fatalf("request SubVersion = %+v", wrapped.LabelRecoveryRequest.Request)
	}

	if got.ShipmentIdentificationNumber != "1ZSHIPMENT" || got.TrackingNumber != "1ZPKG001" {
		t.Fatalf("result ids = %+v", got)
	}
	if got.Label == nil || got.Label.GraphicImage != "R0lGODlh" || got.Label.HTMLImage != "<html/>" {
		t.Fatalf("Label = %+v", got.Label)
	}
}

func TestShipmentShippingLabel_LabelResultsArray(t *testing.T) {
	svc := newTestShipmentService(t, staticJSONHandler(http.StatusOK, `{
		"LabelRecoveryResponse": {
			"Response": {"ResponseStatus": {"Code": "1", "Description": "Success"}},
			"ShipmentIdentificationNumber": "1ZSHIPMENT",
			"LabelResults": [
				{
					"TrackingNumber": "1ZPKG001",
					"LabelImage": {
						"LabelImageFormat": {"Code": "GIF"},
						"GraphicImage": "R0lGODlh"
					}
				}
			]
		}
	}`))

	got, err := svc.ShippingLabel(ctx, ShippingLabelRequest{
		TrackingNumber: "1ZPKG001",
		LabelSpecification: LabelSpecification{
			LabelImageFormat: LabelImageFormat{Code: "GIF"},
		},
	})
	if err != nil {
		t.Fatalf("ShippingLabel error: %v", err)
	}
	if got.TrackingNumber != "1ZPKG001" || got.Label == nil || got.Label.GraphicImage != "R0lGODlh" {
		t.Fatalf("result = %+v", got)
	}
}

func TestShipmentShippingLabel_EmptyLabel(t *testing.T) {
	svc := newTestShipmentService(t, staticJSONHandler(http.StatusOK, `{
		"LabelRecoveryResponse": {
			"Response": {"ResponseStatus": {"Code": "1", "Description": "Success"}},
			"ShipmentIdentificationNumber": "1ZSHIPMENT",
			"LabelResults": {"TrackingNumber": "1ZPKG001"}
		}
	}`))

	_, err := svc.ShippingLabel(ctx, ShippingLabelRequest{
		TrackingNumber: "1ZPKG001",
		LabelSpecification: LabelSpecification{
			LabelImageFormat: LabelImageFormat{Code: "GIF"},
		},
	})
	if err == nil {
		t.Fatal("expected empty label error")
	}
	if !strings.Contains(err.Error(), "面单数据为空") {
		t.Fatalf("error = %q, want 面单数据为空", err.Error())
	}
}

func TestShipmentShippingLabel_NonSuccessStatus(t *testing.T) {
	svc := newTestShipmentService(t, staticJSONHandler(http.StatusOK, `{
		"LabelRecoveryResponse": {
			"Response": {"ResponseStatus": {"Code": "0", "Description": "Label not found"}}
		}
	}`))

	_, err := svc.ShippingLabel(ctx, ShippingLabelRequest{
		TrackingNumber: "1ZPKG001",
		LabelSpecification: LabelSpecification{
			LabelImageFormat: LabelImageFormat{Code: "GIF"},
		},
	})
	if err == nil {
		t.Fatal("expected status error")
	}
	if !strings.Contains(err.Error(), "Label not found") {
		t.Fatalf("error = %q", err.Error())
	}
}

type httpCapture struct {
	method string
	path   string
	query  url.Values
	header http.Header
	body   []byte
}

func newTestShipmentService(t *testing.T, handler http.Handler) shipmentService {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	cfg := &config.Config{
		Version:        "v2409",
		LabelVersion:   "v1",
		TransactionSrc: "ups-go-test",
		ClientID:       "test-client-id",
		ClientSecret:   "test-client-secret",
		AccountNumber:  "XXXXXX",
	}
	svc := &service{
		config:         cfg,
		httpClient:     resty.New().SetBaseURL(srv.URL),
		accessToken:    "test-token",
		tokenExpiresAt: time.Now().Add(time.Hour),
	}
	return shipmentService{service: svc}
}

func captureHandler(cap *httpCapture, status int, body string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cap.method = r.Method
		cap.path = r.URL.Path
		cap.query = r.URL.Query()
		cap.header = r.Header.Clone()
		cap.body, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}
}

func staticJSONHandler(status int, body string) http.HandlerFunc {
	var unused httpCapture
	return captureHandler(&unused, status, body)
}

func validCreateShipmentRequest() CreateShipmentRequest {
	return createShipmentRequest("XXXXXX")
}

func createShipmentRequest(accountNumber string) CreateShipmentRequest {
	return CreateShipmentRequest{
		Shipment: Shipment{
			Description: null.StringFrom("Test Goods"),
			Shipper: Shipper{
				Name:          "Shipper Name",
				ShipperNumber: accountNumber,
				Phone:         &Phone{Number: "1234567890"},
				Address: Address{
					AddressLine:       []string{"123 Main St"},
					City:              "TIMONIUM",
					StateProvinceCode: null.StringFrom("MD"),
					PostalCode:        null.StringFrom("21093"),
					CountryCode:       "US",
				},
			},
			ShipTo: ShipTo{
				Name:  "Consignee Name",
				Phone: &Phone{Number: "1234567890"},
				Address: Address{
					AddressLine:       []string{"456 Oak Ave"},
					City:              "ALPHARETTA",
					StateProvinceCode: null.StringFrom("GA"),
					PostalCode:        null.StringFrom("30005"),
					CountryCode:       "US",
				},
			},
			PaymentInformation: PaymentInformation{
				ShipmentCharge: []ShipmentCharge{
					{
						Type:        "01",
						BillShipper: &BillShipper{AccountNumber: accountNumber},
					},
				},
			},
			Service: ServiceCode{Code: "03", Description: null.StringFrom("Ground")},
			Package: []Package{
				{
					Packaging: Packaging{Code: "02"},
					PackageWeight: &PackageWeight{
						UnitOfMeasurement: UnitOfMeasurement{Code: "LBS"},
						Weight:            "10",
					},
				},
			},
		},
		LabelSpecification: LabelSpecification{
			LabelImageFormat: LabelImageFormat{Code: "GIF"},
			HTTPUserAgent:    null.StringFrom("Mozilla/4.5"),
		},
	}
}
