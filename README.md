# ups-go

UPS 物流 REST API Go SDK（OAuth2 + Ship / Void / Label Recovery）。

## 安装

```bash
go get github.com/hiscaler/ups-go
```

## 配置

```go
cfg := config.Config{
    Debug:          true,
    Env:            "test", // test/dev → CIE；prod → 生产
    Timeout:        30,
    ClientID:       "YOUR_CLIENT_ID",
    ClientSecret:   "YOUR_CLIENT_SECRET",
    AccountNumber:  "XXXXXX",
    TransactionSrc: "ups-go",
    Version:        "v2409",
    LabelVersion:   "v1",
}
client := ups.NewClient(ctx, cfg)
```

本地联调可复制 `config/config.json.example` 为 `config/config.json`（已加入 `.gitignore`）。

## 用法

### 获取 Token

```go
token, err := client.Services.Auth.Token(ctx)
```

业务接口会自动缓存并刷新 Token；一般无需手动调用。

### 发货

```go
result, err := client.Services.Shipment.Create(ctx, ups.CreateShipmentRequest{
    Shipment: ups.Shipment{
        Description: null.StringFrom("Goods"),
        Shipper: ups.Shipper{
            Name:          "Shipper",
            ShipperNumber: "XXXXXX",
            Phone:         &ups.Phone{Number: "1234567890"},
            Address: ups.Address{
                AddressLine:       []string{"123 Main St"},
                City:              "TIMONIUM",
                StateProvinceCode: null.StringFrom("MD"),
                PostalCode:        null.StringFrom("21093"),
                CountryCode:       "US",
            },
        },
        ShipTo: ups.ShipTo{ /* ... */ },
        PaymentInformation: ups.PaymentInformation{
            ShipmentCharge: []ups.ShipmentCharge{{
                Type:        "01",
                BillShipper: &ups.BillShipper{AccountNumber: "XXXXXX"},
            }},
        },
        Service: ups.ServiceCode{Code: "03"},
        Package: []ups.Package{{
            Packaging: ups.Packaging{Code: "02"},
            PackageWeight: &ups.PackageWeight{
                UnitOfMeasurement: ups.UnitOfMeasurement{Code: "LBS"},
                Weight:            "10",
            },
        }},
    },
    LabelSpecification: ups.LabelSpecification{
        LabelImageFormat: ups.LabelImageFormat{Code: "GIF"},
    },
})
```

### 取消发货

```go
voidResult, err := client.Services.Shipment.Cancel(ctx, ups.CancelShipmentRequest{
    ShipmentIdentificationNumber: "1ZXXXXXXXXXXXXXXXX",
})
```

### 获取面单（Label Recovery）

```go
label, err := client.Services.Shipment.ShippingLabel(ctx, ups.ShippingLabelRequest{
    TrackingNumber: "1ZXXXXXXXXXXXXXXXX",
    LabelSpecification: ups.LabelSpecification{
        LabelImageFormat: ups.LabelImageFormat{Code: "GIF"},
    },
})
// label.Label.GraphicImage 为 Base64 面单数据
```

## API 路径

| 能力 | 方法 | 路径 |
|------|------|------|
| OAuth Token | POST | `/security/v1/oauth/token` |
| 发货 | POST | `/api/shipments/{version}/ship` |
| 取消 | DELETE | `/api/shipments/{version}/void/cancel/{shipmentId}` |
| 面单 | POST | `/api/labels/{version}/recovery` |

环境：CIE `https://wwwcie.ups.com`，生产 `https://onlinetools.ups.com`。
