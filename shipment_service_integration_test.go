package ups

import (
	"strings"
	"testing"

	"gopkg.in/guregu/null.v4"
)

func requireIntegrationClient(t *testing.T) {
	t.Helper()
	if client == nil {
		t.Skip("client not initialized: copy config/config.json.example to config/config.json")
	}
	if integrationAccountNumber == "" || integrationAccountNumber == "XXXXXX" {
		t.Skip("account number not configured")
	}
}

func requireIntegrationMutation(t *testing.T) {
	t.Helper()
	requireIntegrationClient(t)
	//if integrationEnv == entity.Prod && os.Getenv("UPS_GO_INTEGRATION") != "1" {
	//	t.Skip("refusing Create/Cancel on prod without UPS_GO_INTEGRATION=1")
	//}
}

func createIntegrationShipmentRequest(accountNumber string) CreateShipmentRequest {
	return CreateShipmentRequest{
		Shipment: Shipment{
			Description: null.StringFrom("Test Goods"),
			Shipper: Shipper{
				Name:          "测试发件人",
				AttentionName: null.StringFrom("测试联系人"),
				ShipperNumber: accountNumber,
				Phone:         &Phone{Number: "4008208388"},
				Address: Address{
					AddressLine:       []string{"浦东新区测试路123号"},
					City:              "SHANGHAI",
					StateProvinceCode: null.StringFrom("SH"),
					PostalCode:        null.StringFrom("200120"),
					CountryCode:       "CN",
				},
			},
			ShipTo: ShipTo{
				Name:          "John",
				AttentionName: null.StringFrom("Consignee Contact"),
				Phone:         &Phone{Number: "1234567890"},
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
			Service: ServiceCode{Code: "07", Description: null.StringFrom("Express")},
			Package: []Package{
				{
					Packaging: Packaging{Code: "02"},
					PackageWeight: &PackageWeight{
						UnitOfMeasurement: UnitOfMeasurement{Code: "KGS"},
						Weight:            "1",
					},
				},
			},
		},
		LabelSpecification: LabelSpecification{
			LabelImageFormat: LabelImageFormat{Code: "PDF"},
		},
	}
}

func TestIntegrationShipmentCreate(t *testing.T) {
	requireIntegrationMutation(t)

	created, err := client.Services.Shipment.Create(ctx, createIntegrationShipmentRequest(integrationAccountNumber))
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}
	if created.ShipmentIdentificationNumber == "" {
		t.Fatal("ShipmentIdentificationNumber is empty")
	}
	if len(created.PackageResults) == 0 {
		t.Fatal("PackageResults is empty")
	}
	pkg := created.PackageResults[0]
	if pkg.TrackingNumber == "" {
		t.Fatal("TrackingNumber is empty")
	}
	if pkg.ShippingLabel == nil || pkg.ShippingLabel.GraphicImage == "" {
		t.Fatal("ShippingLabel is empty")
	}

	t.Logf("shipmentId=%s trackingNumber=%s", created.ShipmentIdentificationNumber, pkg.TrackingNumber)

	t.Cleanup(func() {
		voided, err := client.Services.Shipment.Cancel(ctx, CancelShipmentRequest{
			ShipmentIdentificationNumber: created.ShipmentIdentificationNumber,
		})
		if err != nil {
			t.Errorf("cleanup Cancel error: %v", err)
			return
		}
		t.Logf("cleanup void summary=%s %s", voided.SummaryStatusCode, voided.SummaryStatusDescription)
	})
}

func TestIntegrationShipmentShippingLabel(t *testing.T) {
	requireIntegrationMutation(t)

	created, err := client.Services.Shipment.Create(ctx, createIntegrationShipmentRequest(integrationAccountNumber))
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}
	if len(created.PackageResults) == 0 || created.PackageResults[0].TrackingNumber == "" {
		t.Fatal("missing tracking number from Create")
	}
	trackingNumber := created.PackageResults[0].TrackingNumber
	t.Cleanup(func() {
		_, err = client.Services.Shipment.Cancel(ctx, CancelShipmentRequest{
			ShipmentIdentificationNumber: created.ShipmentIdentificationNumber,
		})
		if err != nil {
			t.Errorf("cleanup Cancel error: %v", err)
		}
	})

	label, err := client.Services.Shipment.ShippingLabel(ctx, ShippingLabelRequest{
		TrackingNumber: trackingNumber,
		LabelSpecification: LabelSpecification{
			LabelImageFormat: LabelImageFormat{Code: "PDF"},
		},
	})
	if err != nil {
		t.Fatalf("ShippingLabel error: %v", err)
	}
	if label.TrackingNumber == "" {
		t.Fatal("TrackingNumber is empty")
	}
	if !strings.EqualFold(label.TrackingNumber, trackingNumber) {
		t.Fatalf("TrackingNumber = %q, want %q", label.TrackingNumber, trackingNumber)
	}
	if label.Label == nil || label.Label.GraphicImage == "" {
		t.Fatal("Label.GraphicImage is empty")
	}
	t.Logf("label shipmentId=%s trackingNumber=%s imageLen=%d",
		label.ShipmentIdentificationNumber, label.TrackingNumber, len(label.Label.GraphicImage))
}

func TestIntegrationShipmentCancel(t *testing.T) {
	requireIntegrationMutation(t)

	created, err := client.Services.Shipment.Create(ctx, createIntegrationShipmentRequest(integrationAccountNumber))
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}
	if created.ShipmentIdentificationNumber == "" {
		t.Fatal("ShipmentIdentificationNumber is empty")
	}

	voided, err := client.Services.Shipment.Cancel(ctx, CancelShipmentRequest{
		ShipmentIdentificationNumber: created.ShipmentIdentificationNumber,
		TrackingNumbers:              []string{created.PackageResults[0].TrackingNumber},
	})
	if err != nil {
		t.Fatalf("Cancel error: %v", err)
	}
	if voided.SummaryStatusCode == "" {
		t.Fatal("SummaryStatusCode is empty")
	}
	if len(voided.PackageLevelResults) == 0 {
		t.Fatal("PackageLevelResults is empty")
	}
	t.Logf("void summary=%s %s package=%s %s",
		voided.SummaryStatusCode, voided.SummaryStatusDescription,
		voided.PackageLevelResults[0].TrackingNumber, voided.PackageLevelResults[0].StatusDescription)
}
