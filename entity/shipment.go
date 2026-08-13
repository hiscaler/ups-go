package entity

// ShipmentResult 发货结果
type ShipmentResult struct {
	ShipmentIdentificationNumber string           `json:"shipmentIdentificationNumber"`    // 运单号
	BillingWeight                *BillingWeight   `json:"billingWeight,omitempty"`         // 计费重量
	ShipmentCharges              *ShipmentCharges `json:"shipmentCharges,omitempty"`       // 运费
	NegotiatedRateCharges        *ShipmentCharges `json:"negotiatedRateCharges,omitempty"` // 协议价
	PackageResults               []PackageResult  `json:"packageResults,omitempty"`        // 包裹结果
}

// BillingWeight 计费重量
type BillingWeight struct {
	UnitOfMeasurement CodeDescription `json:"unitOfMeasurement"`
	Weight            string          `json:"weight"`
}

// ShipmentCharges 运费
type ShipmentCharges struct {
	TransportationCharges *MonetaryAmount `json:"transportationCharges,omitempty"`
	ServiceOptionsCharges *MonetaryAmount `json:"serviceOptionsCharges,omitempty"`
	TotalCharges          *MonetaryAmount `json:"totalCharges,omitempty"`
	BaseServiceCharge     *MonetaryAmount `json:"baseServiceCharge,omitempty"`
}

// MonetaryAmount 金额
type MonetaryAmount struct {
	CurrencyCode  string `json:"currencyCode"`
	MonetaryValue string `json:"monetaryValue"`
}

// PackageResult 包裹结果
type PackageResult struct {
	TrackingNumber        string          `json:"trackingNumber"`                  // 追踪号
	BaseServiceCharge     *MonetaryAmount `json:"baseServiceCharge,omitempty"`     // 基础服务费
	ServiceOptionsCharges *MonetaryAmount `json:"serviceOptionsCharges,omitempty"` // 附加服务费
	ShippingLabel         *ShippingLabel  `json:"shippingLabel,omitempty"`         // 面单
}

// ShippingLabel 面单
type ShippingLabel struct {
	ImageFormat  CodeDescription `json:"imageFormat"`         // 图像格式
	GraphicImage string          `json:"graphicImage"`        // Base64 面单数据
	HTMLImage    string          `json:"htmlImage,omitempty"` // HTML 面单（GIF/PNG）
	URL          string          `json:"url,omitempty"`       // 面单链接
}

// CodeDescription 代码与描述
type CodeDescription struct {
	Code        string `json:"code"`
	Description string `json:"description,omitempty"`
}

// VoidResult 取消发货结果
type VoidResult struct {
	SummaryStatusCode        string              `json:"summaryStatusCode"`
	SummaryStatusDescription string              `json:"summaryStatusDescription"`
	PackageLevelResults      []PackageVoidResult `json:"packageLevelResults,omitempty"`
}

// PackageVoidResult 包裹级取消结果
type PackageVoidResult struct {
	TrackingNumber    string `json:"trackingNumber"`
	StatusCode        string `json:"statusCode"`
	StatusDescription string `json:"statusDescription"`
}

// LabelResult 补打面单结果
type LabelResult struct {
	ShipmentIdentificationNumber string         `json:"shipmentIdentificationNumber"` // 运单号
	TrackingNumber               string         `json:"trackingNumber"`               // 追踪号
	Label                        *ShippingLabel `json:"label,omitempty"`              // 面单数据
}
