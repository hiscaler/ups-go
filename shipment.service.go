package ups

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/hiscaler/ups-go/entity"
	"gopkg.in/guregu/null.v4"
)

// 发货服务
type shipmentService struct {
	*service
}

// ---------- Create Shipment ----------

// CreateShipmentRequest 发货请求（常用子集）
type CreateShipmentRequest struct {
	Request            *ShipmentRequestOption `json:"Request,omitempty"`  // 请求选项（如 SubVersion）
	Shipment           Shipment               `json:"Shipment"`           // 发货信息
	LabelSpecification LabelSpecification     `json:"LabelSpecification"` // 面单规格
}

// Validate 校验发货请求
func (m CreateShipmentRequest) Validate() error {
	return validation.ValidateStruct(&m,
		validation.Field(&m.Shipment),
		validation.Field(&m.LabelSpecification),
	)
}

// ShipmentRequestOption 发货请求选项
type ShipmentRequestOption struct {
	SubVersion           null.String `json:"SubVersion,omitempty"`    // 子版本号
	RequestOption        null.String `json:"RequestOption,omitempty"` // 请求选项
	TransactionReference *struct {
		CustomerContext null.String `json:"CustomerContext,omitempty"`
	} `json:"TransactionReference,omitempty"`
}

// Shipment 发货主体信息
type Shipment struct {
	Description           null.String            `json:"Description,omitempty"`           // 货物描述
	Shipper               Shipper                `json:"Shipper"`                         // 发件人
	ShipTo                ShipTo                 `json:"ShipTo"`                          // 收件人
	ShipFrom              *ShipFrom              `json:"ShipFrom,omitempty"`              // 寄件地址（可与发件人不同）
	PaymentInformation    PaymentInformation     `json:"PaymentInformation"`              // 付款信息
	Service               ServiceCode            `json:"Service"`                         // 服务类型
	Package               []Package              `json:"Package"`                         // 包裹列表
	ShipmentRatingOptions *ShipmentRatingOptions `json:"ShipmentRatingOptions,omitempty"` // 费率选项
}

// Validate 校验发货主体信息
func (m Shipment) Validate() error {
	return validation.ValidateStruct(&m,
		validation.Field(&m.Description, validation.When(m.Description.Valid, validation.Length(1, 50).Error("货物描述长度必须在 {{.min}}-{{.max}} 之间"))),
		validation.Field(&m.Shipper),
		validation.Field(&m.ShipTo),
		validation.Field(&m.PaymentInformation),
		validation.Field(&m.Service),
		validation.Field(&m.Package, validation.Required.Error("包裹信息不能为空"), validation.Length(1, 200).Error("包裹数量必须在 {{.min}}-{{.max}} 之间")),
	)
}

// Shipper 发件人信息
type Shipper struct {
	Name                    string      `json:"Name"`                              // 发件人姓名
	AttentionName           null.String `json:"AttentionName,omitempty"`           // 联系人
	TaxIdentificationNumber null.String `json:"TaxIdentificationNumber,omitempty"` // 税号
	Phone                   *Phone      `json:"Phone,omitempty"`                   // 电话
	ShipperNumber           string      `json:"ShipperNumber"`                     // UPS 账号（6 位）
	FaxNumber               null.String `json:"FaxNumber,omitempty"`               // 传真
	EMailAddress            null.String `json:"EMailAddress,omitempty"`            // 邮箱
	Address                 Address     `json:"Address"`                           // 地址
}

// Validate 校验发件人信息
func (m Shipper) Validate() error {
	return validation.ValidateStruct(&m,
		validation.Field(&m.Name, validation.Required.Error("发件人姓名不能为空"), validation.Length(1, 35).Error("发件人姓名长度必须在 {{.min}}-{{.max}} 之间")),
		validation.Field(&m.ShipperNumber, validation.Required.Error("发件人账号不能为空"), validation.Length(6, 6).Error("发件人账号长度必须为 {{.min}} 位")),
		validation.Field(&m.Address),
	)
}

// ShipTo 收件人信息
type ShipTo struct {
	Name                    string      `json:"Name"`                              // 收件人姓名
	AttentionName           null.String `json:"AttentionName,omitempty"`           // 联系人
	Phone                   *Phone      `json:"Phone,omitempty"`                   // 电话
	FaxNumber               null.String `json:"FaxNumber,omitempty"`               // 传真
	TaxIdentificationNumber null.String `json:"TaxIdentificationNumber,omitempty"` // 税号
	EMailAddress            null.String `json:"EMailAddress,omitempty"`            // 邮箱
	Address                 Address     `json:"Address"`                           // 地址
}

// Validate 校验收件人信息
func (m ShipTo) Validate() error {
	return validation.ValidateStruct(&m,
		validation.Field(&m.Name, validation.Required.Error("收件人姓名不能为空"), validation.Length(1, 35).Error("收件人姓名长度必须在 {{.min}}-{{.max}} 之间")),
		validation.Field(&m.Address),
	)
}

// ShipFrom 寄件地址信息
type ShipFrom struct {
	Name                    string      `json:"Name"`                              // 姓名
	AttentionName           null.String `json:"AttentionName,omitempty"`           // 联系人
	Phone                   *Phone      `json:"Phone,omitempty"`                   // 电话
	FaxNumber               null.String `json:"FaxNumber,omitempty"`               // 传真
	TaxIdentificationNumber null.String `json:"TaxIdentificationNumber,omitempty"` // 税号
	EMailAddress            null.String `json:"EMailAddress,omitempty"`            // 邮箱
	Address                 Address     `json:"Address"`                           // 地址
}

// Validate 校验寄件地址信息
func (m ShipFrom) Validate() error {
	return validation.ValidateStruct(&m,
		validation.Field(&m.Name, validation.Required.Error("寄件地址姓名不能为空"), validation.Length(1, 35).Error("寄件地址姓名长度必须在 {{.min}}-{{.max}} 之间")),
		validation.Field(&m.Address),
	)
}

// Phone 电话信息
type Phone struct {
	Number    string      `json:"Number"`              // 电话号码
	Extension null.String `json:"Extension,omitempty"` // 分机号
}

// Validate 校验电话信息
func (m Phone) Validate() error {
	return validation.ValidateStruct(&m,
		validation.Field(&m.Number, validation.Required.Error("电话号码不能为空"), validation.Length(1, 15).Error("电话号码长度必须在 {{.min}}-{{.max}} 之间")),
	)
}

// Address 地址信息
type Address struct {
	AddressLine       []string    `json:"AddressLine"`                 // 地址行（最多 3 行）
	City              string      `json:"City"`                        // 城市
	StateProvinceCode null.String `json:"StateProvinceCode,omitempty"` // 州/省代码
	PostalCode        null.String `json:"PostalCode,omitempty"`        // 邮编
	CountryCode       string      `json:"CountryCode"`                 // 国家代码（2 位）
}

// Validate 校验地址信息
func (m Address) Validate() error {
	return validation.ValidateStruct(&m,
		validation.Field(&m.AddressLine, validation.Required.Error("地址行不能为空"), validation.Length(1, 3).Error("地址行数量必须在 {{.min}}-{{.max}} 之间")),
		validation.Field(&m.City, validation.Required.Error("城市不能为空"), validation.Length(1, 30).Error("城市长度必须在 {{.min}}-{{.max}} 之间")),
		validation.Field(&m.CountryCode, validation.Required.Error("国家代码不能为空"), validation.Length(2, 2).Error("国家代码长度必须为 {{.min}} 位")),
	)
}

// PaymentInformation 付款信息
type PaymentInformation struct {
	ShipmentCharge []ShipmentCharge `json:"ShipmentCharge"` // 费用项（最多 3 项）
}

// Validate 校验付款信息
func (m PaymentInformation) Validate() error {
	return validation.ValidateStruct(&m,
		validation.Field(&m.ShipmentCharge, validation.Required.Error("付款信息不能为空"), validation.Length(1, 3).Error("付款项数量必须在 {{.min}}-{{.max}} 之间")),
	)
}

// ShipmentCharge 发货费用项
type ShipmentCharge struct {
	Type        string       `json:"Type"`                  // 01=运输费, 02=关税税费, 03=Broker of Choice
	BillShipper *BillShipper `json:"BillShipper,omitempty"` // 发件人付款
}

// Validate 校验费用项
func (m ShipmentCharge) Validate() error {
	return validation.ValidateStruct(&m,
		validation.Field(&m.Type, validation.Required.Error("付款类型不能为空"), validation.In("01", "02", "03").Error("付款类型只能为 01、02 或 03")),
		validation.Field(&m.BillShipper, validation.When(m.Type == "01", validation.Required.Error("运输费付款方不能为空"))),
	)
}

// BillShipper 发件人付款账户
type BillShipper struct {
	AccountNumber string `json:"AccountNumber"` // UPS 付款账号（6 位）
}

// Validate 校验付款账号
func (m BillShipper) Validate() error {
	return validation.ValidateStruct(&m,
		validation.Field(&m.AccountNumber, validation.Required.Error("付款账号不能为空"), validation.Length(6, 6).Error("付款账号长度必须为 {{.min}} 位")),
	)
}

// ServiceCode UPS 服务代码
type ServiceCode struct {
	Code        string      `json:"Code"`                  // 服务代码（2 位，如 03=Ground）
	Description null.String `json:"Description,omitempty"` // 服务描述
}

// Validate 校验服务代码
func (m ServiceCode) Validate() error {
	return validation.ValidateStruct(&m,
		validation.Field(&m.Code, validation.Required.Error("服务代码不能为空"), validation.Length(2, 2).Error("服务代码长度必须为 {{.min}} 位")),
	)
}

// Package 包裹信息
type Package struct {
	Description   null.String    `json:"Description,omitempty"`   // 包裹描述
	Packaging     Packaging      `json:"Packaging"`               // 包装类型
	Dimensions    *Dimensions    `json:"Dimensions,omitempty"`    // 尺寸
	PackageWeight *PackageWeight `json:"PackageWeight,omitempty"` // 重量
}

// Validate 校验包裹信息
func (m Package) Validate() error {
	return validation.ValidateStruct(&m,
		validation.Field(&m.Packaging),
		validation.Field(&m.PackageWeight, validation.Required.Error("包裹重量不能为空")),
	)
}

// Packaging 包装类型
type Packaging struct {
	Code        string      `json:"Code"`                  // 包装代码（如 02=Customer Supplied）
	Description null.String `json:"Description,omitempty"` // 包装描述
}

// Validate 校验包装类型
func (m Packaging) Validate() error {
	return validation.ValidateStruct(&m,
		validation.Field(&m.Code, validation.Required.Error("包装类型不能为空"), validation.Length(2, 2).Error("包装类型长度必须为 {{.min}} 位")),
	)
}

// Dimensions 包裹尺寸
type Dimensions struct {
	UnitOfMeasurement UnitOfMeasurement `json:"UnitOfMeasurement"` // 长度单位（IN/CM）
	Length            string            `json:"Length"`            // 长
	Width             string            `json:"Width"`             // 宽
	Height            string            `json:"Height"`            // 高
}

// Validate 校验包裹尺寸
func (m Dimensions) Validate() error {
	return validation.ValidateStruct(&m,
		validation.Field(&m.UnitOfMeasurement),
		validation.Field(&m.Length, validation.Required.Error("长度不能为空")),
		validation.Field(&m.Width, validation.Required.Error("宽度不能为空")),
		validation.Field(&m.Height, validation.Required.Error("高度不能为空")),
	)
}

// PackageWeight 包裹重量
type PackageWeight struct {
	UnitOfMeasurement UnitOfMeasurement `json:"UnitOfMeasurement"` // 重量单位（LBS/KGS）
	Weight            string            `json:"Weight"`            // 重量
}

// Validate 校验包裹重量
func (m PackageWeight) Validate() error {
	return validation.ValidateStruct(&m,
		validation.Field(&m.UnitOfMeasurement),
		validation.Field(&m.Weight, validation.Required.Error("重量不能为空")),
	)
}

// UnitOfMeasurement 计量单位
type UnitOfMeasurement struct {
	Code        string      `json:"Code"`                  // 单位代码
	Description null.String `json:"Description,omitempty"` // 单位描述
}

// Validate 校验计量单位
func (m UnitOfMeasurement) Validate() error {
	return validation.ValidateStruct(&m,
		validation.Field(&m.Code, validation.Required.Error("计量单位不能为空")),
	)
}

// ShipmentRatingOptions 费率选项
type ShipmentRatingOptions struct {
	NegotiatedRatesIndicator null.String `json:"NegotiatedRatesIndicator,omitempty"` // 协议价标识（空字符串即可）
}

// LabelSpecification 面单规格
type LabelSpecification struct {
	LabelImageFormat LabelImageFormat `json:"LabelImageFormat"`         // 面单图像格式
	HTTPUserAgent    null.String      `json:"HTTPUserAgent,omitempty"`  // GIF 时建议填写
	LabelStockSize   *LabelStockSize  `json:"LabelStockSize,omitempty"` // 热敏面单纸张尺寸
	CharacterSet     null.String      `json:"CharacterSet,omitempty"`   // 字符集
}

// Validate 校验面单规格
func (m LabelSpecification) Validate() error {
	code := strings.ToUpper(m.LabelImageFormat.Code)
	return validation.ValidateStruct(&m,
		validation.Field(&m.LabelImageFormat),
		validation.Field(&m.LabelStockSize, validation.When(
			code == "ZPL" || code == "EPL" || code == "SPL",
			validation.Required.Error("热敏面单必须指定 LabelStockSize"),
		)),
	)
}

// LabelImageFormat 面单图像格式
type LabelImageFormat struct {
	Code        string      `json:"Code"`                  // GIF/PNG/ZPL/EPL/SPL/PDF
	Description null.String `json:"Description,omitempty"` // 格式描述
}

// Validate 校验面单图像格式
func (m LabelImageFormat) Validate() error {
	return validation.ValidateStruct(&m,
		validation.Field(&m.Code,
			validation.Required.Error("面单格式不能为空"),
			validation.In("GIF", "PNG", "ZPL", "EPL", "SPL", "PDF", "gif", "png", "zpl", "epl", "spl", "pdf").
				Error("面单格式只能为 GIF、PNG、ZPL、EPL、SPL 或 PDF")),
	)
}

// LabelStockSize 热敏面单纸张尺寸
type LabelStockSize struct {
	Height string `json:"Height"` // 高度：6 或 8
	Width  string `json:"Width"`  // 宽度：4
}

// Validate 校验面单纸张尺寸
func (m LabelStockSize) Validate() error {
	return validation.ValidateStruct(&m,
		validation.Field(&m.Height, validation.Required.Error("面单高度不能为空"), validation.In("6", "8").Error("面单高度只能为 6 或 8")),
		validation.Field(&m.Width, validation.Required.Error("面单宽度不能为空"), validation.In("4").Error("面单宽度只能为 4")),
	)
}

type shipmentRequestWrapper struct {
	ShipmentRequest CreateShipmentRequest `json:"ShipmentRequest"`
}

type shipmentResponseWrapper struct {
	ShipmentResponse struct {
		Response struct {
			ResponseStatus struct {
				Code        string `json:"Code"`
				Description string `json:"Description"`
			} `json:"ResponseStatus"`
		} `json:"Response"`
		ShipmentResults rawShipmentResults `json:"ShipmentResults"`
	} `json:"ShipmentResponse"`
}

type rawShipmentResults struct {
	ShipmentIdentificationNumber string              `json:"ShipmentIdentificationNumber"`
	BillingWeight                *rawBillingWeight   `json:"BillingWeight"`
	ShipmentCharges              *rawShipmentCharges `json:"ShipmentCharges"`
	NegotiatedRateCharges        *rawShipmentCharges `json:"NegotiatedRateCharges"`
	PackageResults               json.RawMessage     `json:"PackageResults"`
}

type rawBillingWeight struct {
	UnitOfMeasurement entity.CodeDescription `json:"UnitOfMeasurement"`
	Weight            string                 `json:"Weight"`
}

type rawShipmentCharges struct {
	TransportationCharges *rawMonetary `json:"TransportationCharges"`
	ServiceOptionsCharges *rawMonetary `json:"ServiceOptionsCharges"`
	TotalCharges          *rawMonetary `json:"TotalCharges"`
	BaseServiceCharge     *rawMonetary `json:"BaseServiceCharge"`
}

type rawMonetary struct {
	CurrencyCode  string `json:"CurrencyCode"`
	MonetaryValue string `json:"MonetaryValue"`
}

type rawPackageResult struct {
	TrackingNumber        string            `json:"TrackingNumber"`
	BaseServiceCharge     *rawMonetary      `json:"BaseServiceCharge"`
	ServiceOptionsCharges *rawMonetary      `json:"ServiceOptionsCharges"`
	ShippingLabel         *rawShippingLabel `json:"ShippingLabel"`
}

type rawShippingLabel struct {
	ImageFormat  entity.CodeDescription `json:"ImageFormat"`
	GraphicImage string                 `json:"GraphicImage"`
	HTMLImage    string                 `json:"HTMLImage"`
	URL          string                 `json:"URL"`
}

// Create 创建发货
func (s shipmentService) Create(ctx context.Context, req CreateShipmentRequest) (entity.ShipmentResult, error) {
	if err := req.Validate(); err != nil {
		return entity.ShipmentResult{}, invalidInput(err)
	}

	var result entity.ShipmentResult
	err := s.doWithAuth(ctx, func(ctx context.Context, token string) error {
		var res shipmentResponseWrapper
		path := fmt.Sprintf("/api/shipments/%s/ship", s.apiVersion())
		r := s.withAuthHeaders(s.httpClient.R().SetContext(ctx), token).
			SetBody(shipmentRequestWrapper{ShipmentRequest: req}).
			SetQueryParam("additionaladdressvalidation", "city").
			SetResult(&res)
		resp, e := r.Post(path)
		if e = recheckError(resp, e); e != nil {
			return e
		}
		status := res.ShipmentResponse.Response.ResponseStatus.Code
		if status != "" && status != "1" {
			return errorWrap(status, res.ShipmentResponse.Response.ResponseStatus.Description)
		}
		parsed, e := mapShipmentResult(res.ShipmentResponse.ShipmentResults)
		if e != nil {
			return e
		}
		result = parsed
		return nil
	})
	return result, err
}

// mapShipmentResult 将 UPS 原始发货结果映射为实体
func mapShipmentResult(raw rawShipmentResults) (entity.ShipmentResult, error) {
	out := entity.ShipmentResult{
		ShipmentIdentificationNumber: raw.ShipmentIdentificationNumber,
	}
	if raw.BillingWeight != nil {
		out.BillingWeight = &entity.BillingWeight{
			UnitOfMeasurement: raw.BillingWeight.UnitOfMeasurement,
			Weight:            raw.BillingWeight.Weight,
		}
	}
	out.ShipmentCharges = mapCharges(raw.ShipmentCharges)
	out.NegotiatedRateCharges = mapCharges(raw.NegotiatedRateCharges)

	pkgs, err := unmarshalPackageResults(raw.PackageResults)
	if err != nil {
		return entity.ShipmentResult{}, err
	}
	out.PackageResults = make([]entity.PackageResult, 0, len(pkgs))
	for _, p := range pkgs {
		pr := entity.PackageResult{
			TrackingNumber:        p.TrackingNumber,
			BaseServiceCharge:     mapMonetary(p.BaseServiceCharge),
			ServiceOptionsCharges: mapMonetary(p.ServiceOptionsCharges),
		}
		if p.ShippingLabel != nil {
			pr.ShippingLabel = &entity.ShippingLabel{
				ImageFormat:  p.ShippingLabel.ImageFormat,
				GraphicImage: p.ShippingLabel.GraphicImage,
				HTMLImage:    p.ShippingLabel.HTMLImage,
				URL:          p.ShippingLabel.URL,
			}
		}
		out.PackageResults = append(out.PackageResults, pr)
	}
	return out, nil
}

// mapCharges 映射运费结构
func mapCharges(raw *rawShipmentCharges) *entity.ShipmentCharges {
	if raw == nil {
		return nil
	}
	return &entity.ShipmentCharges{
		TransportationCharges: mapMonetary(raw.TransportationCharges),
		ServiceOptionsCharges: mapMonetary(raw.ServiceOptionsCharges),
		TotalCharges:          mapMonetary(raw.TotalCharges),
		BaseServiceCharge:     mapMonetary(raw.BaseServiceCharge),
	}
}

// mapMonetary 映射金额结构
func mapMonetary(raw *rawMonetary) *entity.MonetaryAmount {
	if raw == nil {
		return nil
	}
	return &entity.MonetaryAmount{
		CurrencyCode:  raw.CurrencyCode,
		MonetaryValue: raw.MonetaryValue,
	}
}

// unmarshalPackageResults 解析 PackageResults（兼容单对象与数组）
func unmarshalPackageResults(raw json.RawMessage) ([]rawPackageResult, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return nil, nil
	}
	raw = json.RawMessage(strings.TrimSpace(string(raw)))
	if len(raw) > 0 && raw[0] == '{' {
		var one rawPackageResult
		if err := json.Unmarshal(raw, &one); err != nil {
			return nil, err
		}
		return []rawPackageResult{one}, nil
	}
	var many []rawPackageResult
	if err := json.Unmarshal(raw, &many); err != nil {
		return nil, err
	}
	return many, nil
}

// ---------- Cancel / Void ----------

// CancelShipmentRequest 取消发货请求
type CancelShipmentRequest struct {
	ShipmentIdentificationNumber string   // 运单号（ShipmentIdentificationNumber）
	TrackingNumbers              []string // 可选，包裹级取消（最多 20 个）
}

// Validate 校验取消发货请求
func (m CancelShipmentRequest) Validate() error {
	return validation.ValidateStruct(&m,
		validation.Field(&m.ShipmentIdentificationNumber,
			validation.Required.Error("运单号不能为空"),
			validation.Length(1, 18).Error("运单号长度必须在 {{.min}}-{{.max}} 之间")),
		validation.Field(&m.TrackingNumbers, validation.Length(0, 20).Error("追踪号数量不能超过 {{.max}}")),
	)
}

type voidResponseWrapper struct {
	VoidShipmentResponse struct {
		Response struct {
			ResponseStatus struct {
				Code        string `json:"Code"`
				Description string `json:"Description"`
			} `json:"ResponseStatus"`
		} `json:"Response"`
		SummaryResult struct {
			Status struct {
				Code        string `json:"Code"`
				Description string `json:"Description"`
			} `json:"Status"`
		} `json:"SummaryResult"`
		PackageLevelResult json.RawMessage `json:"PackageLevelResult"`
	} `json:"VoidShipmentResponse"`
}

type rawPackageVoidResult struct {
	TrackingNumber string `json:"TrackingNumber"`
	Status         struct {
		Code        string `json:"Code"`
		Description string `json:"Description"`
	} `json:"Status"`
}

// Cancel 取消发货（Void）
func (s shipmentService) Cancel(ctx context.Context, req CancelShipmentRequest) (entity.VoidResult, error) {
	if err := req.Validate(); err != nil {
		return entity.VoidResult{}, invalidInput(err)
	}

	var result entity.VoidResult
	err := s.doWithAuth(ctx, func(ctx context.Context, token string) error {
		var res voidResponseWrapper
		path := fmt.Sprintf("/api/shipments/%s/void/cancel/%s", s.apiVersion(), strings.ToUpper(req.ShipmentIdentificationNumber))
		r := s.withAuthHeaders(s.httpClient.R().SetContext(ctx), token).SetResult(&res)
		if len(req.TrackingNumbers) == 1 {
			r.SetQueryParam("trackingnumber", strings.ToUpper(req.TrackingNumbers[0]))
		} else if len(req.TrackingNumbers) > 1 {
			quoted := make([]string, 0, len(req.TrackingNumbers))
			for _, tn := range req.TrackingNumbers {
				quoted = append(quoted, `"`+strings.ToUpper(tn)+`"`)
			}
			r.SetQueryParam("trackingnumber", "["+strings.Join(quoted, ",")+"]")
		}
		resp, e := r.Delete(path)
		if e = recheckError(resp, e); e != nil {
			return e
		}
		status := res.VoidShipmentResponse.Response.ResponseStatus.Code
		if status != "" && status != "1" {
			return errorWrap(status, res.VoidShipmentResponse.Response.ResponseStatus.Description)
		}
		result = entity.VoidResult{
			SummaryStatusCode:        res.VoidShipmentResponse.SummaryResult.Status.Code,
			SummaryStatusDescription: res.VoidShipmentResponse.SummaryResult.Status.Description,
		}
		pkgs, e := unmarshalPackageVoidResults(res.VoidShipmentResponse.PackageLevelResult)
		if e != nil {
			return e
		}
		for _, p := range pkgs {
			result.PackageLevelResults = append(result.PackageLevelResults, entity.PackageVoidResult{
				TrackingNumber:    p.TrackingNumber,
				StatusCode:        p.Status.Code,
				StatusDescription: p.Status.Description,
			})
		}
		return nil
	})
	return result, err
}

// unmarshalPackageVoidResults 解析包裹级取消结果（兼容单对象与数组）
func unmarshalPackageVoidResults(raw json.RawMessage) ([]rawPackageVoidResult, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return nil, nil
	}
	raw = json.RawMessage(strings.TrimSpace(string(raw)))
	if len(raw) > 0 && raw[0] == '{' {
		var one rawPackageVoidResult
		if err := json.Unmarshal(raw, &one); err != nil {
			return nil, err
		}
		return []rawPackageVoidResult{one}, nil
	}
	var many []rawPackageVoidResult
	if err := json.Unmarshal(raw, &many); err != nil {
		return nil, err
	}
	return many, nil
}

// ---------- Label Recovery ----------

// ShippingLabelRequest 获取/补打面单请求
type ShippingLabelRequest struct {
	TrackingNumber     string             `json:"TrackingNumber"`     // 追踪号
	LabelSpecification LabelSpecification `json:"LabelSpecification"` // 面单规格
	SubVersion         null.String        `json:"-"`                  // 可选子版本
}

// Validate 校验面单请求
func (m ShippingLabelRequest) Validate() error {
	return validation.ValidateStruct(&m,
		validation.Field(&m.TrackingNumber, validation.Required.Error("追踪号不能为空")),
		validation.Field(&m.LabelSpecification),
	)
}

type labelRecoveryRequestWrapper struct {
	LabelRecoveryRequest struct {
		LabelSpecification LabelSpecification `json:"LabelSpecification"`
		Request            *struct {
			SubVersion null.String `json:"SubVersion,omitempty"`
		} `json:"Request,omitempty"`
		TrackingNumber string `json:"TrackingNumber"`
	} `json:"LabelRecoveryRequest"`
}

type labelRecoveryResponseWrapper struct {
	LabelRecoveryResponse struct {
		Response struct {
			ResponseStatus struct {
				Code        string `json:"Code"`
				Description string `json:"Description"`
			} `json:"ResponseStatus"`
		} `json:"Response"`
		ShipmentIdentificationNumber string          `json:"ShipmentIdentificationNumber"`
		LabelResults                 json.RawMessage `json:"LabelResults"`
	} `json:"LabelRecoveryResponse"`
}

type rawLabelResult struct {
	TrackingNumber string `json:"TrackingNumber"`
	LabelImage     *struct {
		LabelImageFormat entity.CodeDescription `json:"LabelImageFormat"`
		GraphicImage     string                 `json:"GraphicImage"`
		HTMLImage        string                 `json:"HTMLImage"`
		URL              string                 `json:"URL"`
	} `json:"LabelImage"`
}

// ShippingLabel 补打/获取面单（Label Recovery）
func (s shipmentService) ShippingLabel(ctx context.Context, req ShippingLabelRequest) (entity.LabelResult, error) {
	if err := req.Validate(); err != nil {
		return entity.LabelResult{}, invalidInput(err)
	}

	var result entity.LabelResult
	err := s.doWithAuth(ctx, func(ctx context.Context, token string) error {
		body := labelRecoveryRequestWrapper{}
		body.LabelRecoveryRequest.LabelSpecification = req.LabelSpecification
		body.LabelRecoveryRequest.TrackingNumber = strings.ToUpper(req.TrackingNumber)
		if req.SubVersion.Valid {
			body.LabelRecoveryRequest.Request = &struct {
				SubVersion null.String `json:"SubVersion,omitempty"`
			}{SubVersion: req.SubVersion}
		}

		var res labelRecoveryResponseWrapper
		path := fmt.Sprintf("/api/labels/%s/recovery", s.labelAPIVersion())
		resp, e := s.withAuthHeaders(s.httpClient.R().SetContext(ctx), token).
			SetBody(body).
			SetResult(&res).
			Post(path)
		if e = recheckError(resp, e); e != nil {
			return e
		}
		status := res.LabelRecoveryResponse.Response.ResponseStatus.Code
		if status != "" && status != "1" {
			return errorWrap(status, res.LabelRecoveryResponse.Response.ResponseStatus.Description)
		}

		labels, e := unmarshalLabelResults(res.LabelRecoveryResponse.LabelResults)
		if e != nil {
			return e
		}
		result = entity.LabelResult{
			ShipmentIdentificationNumber: res.LabelRecoveryResponse.ShipmentIdentificationNumber,
		}
		if len(labels) > 0 {
			result.TrackingNumber = labels[0].TrackingNumber
			if labels[0].LabelImage != nil {
				result.Label = &entity.ShippingLabel{
					ImageFormat:  labels[0].LabelImage.LabelImageFormat,
					GraphicImage: labels[0].LabelImage.GraphicImage,
					HTMLImage:    labels[0].LabelImage.HTMLImage,
					URL:          labels[0].LabelImage.URL,
				}
			}
		}
		if result.Label == nil || result.Label.GraphicImage == "" {
			return errorWrap("", "面单数据为空")
		}
		return nil
	})
	return result, err
}

// unmarshalLabelResults 解析 LabelResults（兼容单对象与数组）
func unmarshalLabelResults(raw json.RawMessage) ([]rawLabelResult, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return nil, nil
	}
	raw = json.RawMessage(strings.TrimSpace(string(raw)))
	if len(raw) > 0 && raw[0] == '{' {
		var one rawLabelResult
		if err := json.Unmarshal(raw, &one); err != nil {
			return nil, err
		}
		return []rawLabelResult{one}, nil
	}
	var many []rawLabelResult
	if err := json.Unmarshal(raw, &many); err != nil {
		return nil, err
	}
	return many, nil
}
