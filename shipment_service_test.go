package ups

import (
	"testing"

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
	req := CreateShipmentRequest{
		Shipment: Shipment{
			Description: null.StringFrom("Test Goods"),
			Shipper: Shipper{
				Name:          "Shipper Name",
				ShipperNumber: "XXXXXX",
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
						BillShipper: &BillShipper{AccountNumber: "XXXXXX"},
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
	if err := req.Validate(); err != nil {
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
