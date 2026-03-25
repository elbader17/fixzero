package fixzero

import (
	"strconv"
)

type MsgType interface {
	MsgType() string
	Encode() []byte
	Decode(data []byte) error
}

type MsgTypeRegistry struct {
	types map[string]func() MsgType
}

var GlobalMsgTypeRegistry = &MsgTypeRegistry{
	types: make(map[string]func() MsgType),
}

func (r *MsgTypeRegistry) Register(msgType string, factory func() MsgType) {
	r.types[msgType] = factory
}

func (r *MsgTypeRegistry) Create(msgType string) MsgType {
	if f, ok := r.types[msgType]; ok {
		return f()
	}
	return nil
}

type Header struct {
	BeginString     string
	BodyLength      int
	MsgType         string
	SenderCompID    string
	TargetCompID    string
	MsgSeqNum       int
	SendingTime     string
	PossDupFlag     bool
	OrigSendingTime string
}

type Trailer struct {
	SignatureLength int
	Signature       string
	CheckSum        string
}

type NewOrderSingleMsg struct {
	Header
	Body    NewOrderSingleBody
	Trailer Trailer
}

type NewOrderSingleBody struct {
	ClOrdID           string
	ClientID          string
	ExecBroker        string
	Account           string
	QuoteReqID        string
	Symbol            string
	SymbolSfx         string
	SecurityID        string
	SecurityIDSource  string
	SecurityAltID     string
	Product           int
	CFICode           string
	SecurityType      string
	MaturityMonthYear string
	MaturityDate      string
	CouponRate        float64
	SecurityDesc      string
	Side              string
	TransactTime      string
	HandlInst         string
	OrdType           string
	Price             float64
	StopPx            float64
	Currency          string
	TimeInForce       string
	OrderQty          float64
	CashOrderQty      float64
	PeggedRef         string
	DiscloseFlag      string
	ManualOrder       bool
}

func (m *NewOrderSingleMsg) MsgType() string {
	return MsgTypeNewOrderSingle
}

func (m *NewOrderSingleMsg) Encode() []byte {
	buf := GetBuffer(512)
	defer PutBuffer(buf)

	buf = appendField(buf, TagBeginString, m.Header.BeginString)
	buf = appendField(buf, TagMsgType, MsgTypeNewOrderSingle)
	buf = appendField(buf, TagSenderCompID, m.Header.SenderCompID)
	buf = appendField(buf, TagTargetCompID, m.Header.TargetCompID)
	buf = appendField(buf, TagMsgSeqNum, formatInt(int64(m.Header.MsgSeqNum)))
	buf = appendField(buf, TagSendingTime, m.Header.SendingTime)
	if m.Header.PossDupFlag {
		buf = appendField(buf, TagPossDupFlag, "Y")
	}

	buf = appendField(buf, TagClOrdID, m.Body.ClOrdID)
	buf = appendField(buf, TagClientID, m.Body.ClientID)
	buf = appendField(buf, TagExecBroker, m.Body.ExecBroker)
	buf = appendField(buf, TagAccount, m.Body.Account)
	buf = appendField(buf, TagQuoteReqID, m.Body.QuoteReqID)
	buf = appendField(buf, TagSymbol, m.Body.Symbol)
	buf = appendField(buf, TagSymbolSfx, m.Body.SymbolSfx)
	buf = appendField(buf, TagSecurityID, m.Body.SecurityID)
	buf = appendField(buf, TagSecurityIDSource, m.Body.SecurityIDSource)
	buf = appendField(buf, TagSecurityAltID, m.Body.SecurityAltID)
	if m.Body.Product > 0 {
		buf = appendField(buf, TagProduct, formatInt(int64(m.Body.Product)))
	}
	buf = appendField(buf, TagCFICode, m.Body.CFICode)
	buf = appendField(buf, TagSecurityType, m.Body.SecurityType)
	buf = appendField(buf, TagMaturityMonthYear, m.Body.MaturityMonthYear)
	buf = appendField(buf, TagMaturityDate, m.Body.MaturityDate)
	if m.Body.CouponRate != 0 {
		buf = appendField(buf, TagCouponRate, strconv.FormatFloat(m.Body.CouponRate, 'f', -1, 64))
	}
	buf = appendField(buf, TagSecurityDesc, m.Body.SecurityDesc)
	buf = appendField(buf, TagSide, m.Body.Side)
	buf = appendField(buf, TagTransactTime, m.Body.TransactTime)
	buf = appendField(buf, TagHandlInst, m.Body.HandlInst)
	buf = appendField(buf, TagOrdType, m.Body.OrdType)
	if m.Body.Price != 0 {
		buf = appendField(buf, TagPrice, strconv.FormatFloat(m.Body.Price, 'f', -1, 64))
	}
	if m.Body.StopPx != 0 {
		buf = appendField(buf, TagStopPx, strconv.FormatFloat(m.Body.StopPx, 'f', -1, 64))
	}
	buf = appendField(buf, TagCurrency, m.Body.Currency)
	buf = appendField(buf, TagTimeInForce, m.Body.TimeInForce)
	if m.Body.OrderQty != 0 {
		buf = appendField(buf, TagQuantity, strconv.FormatFloat(m.Body.OrderQty, 'f', -1, 64))
	}
	if m.Body.CashOrderQty != 0 {
		buf = appendField(buf, TagCashOrderQty, strconv.FormatFloat(m.Body.CashOrderQty, 'f', -1, 64))
	}
	buf = appendField(buf, TagPeggedRef, m.Body.PeggedRef)
	buf = appendField(buf, TagDiscloseFlag, m.Body.DiscloseFlag)
	if m.Body.ManualOrder {
		buf = appendField(buf, TagManualOrder, "Y")
	}

	return serializeWithChecksum(buf)
}

func (m *NewOrderSingleMsg) Decode(data []byte) error {
	msg := GetMessage()
	defer PutMessage(msg)

	if err := parseMessage(data, msg); err != nil {
		return err
	}

	m.Header.BeginString = msg.GetString(TagBeginString)
	m.Header.MsgType = msg.GetString(TagMsgType)
	m.Header.SenderCompID = msg.GetString(TagSenderCompID)
	m.Header.TargetCompID = msg.GetString(TagTargetCompID)
	if seq, ok := msg.GetInt(TagMsgSeqNum); ok {
		m.Header.MsgSeqNum = int(seq)
	}
	m.Header.SendingTime = msg.GetString(TagSendingTime)
	if dup, ok := msg.GetBool(TagPossDupFlag); ok {
		m.Header.PossDupFlag = dup
	}

	m.Body.ClOrdID = msg.GetString(TagClOrdID)
	m.Body.ClientID = msg.GetString(TagClientID)
	m.Body.ExecBroker = msg.GetString(TagExecBroker)
	m.Body.Account = msg.GetString(TagAccount)
	m.Body.QuoteReqID = msg.GetString(TagQuoteReqID)
	m.Body.Symbol = msg.GetString(TagSymbol)
	m.Body.SymbolSfx = msg.GetString(TagSymbolSfx)
	m.Body.SecurityID = msg.GetString(TagSecurityID)
	m.Body.SecurityIDSource = msg.GetString(TagSecurityIDSource)
	m.Body.SecurityAltID = msg.GetString(TagSecurityAltID)
	if p, ok := msg.GetInt(TagProduct); ok {
		m.Body.Product = int(p)
	}
	m.Body.CFICode = msg.GetString(TagCFICode)
	m.Body.SecurityType = msg.GetString(TagSecurityType)
	m.Body.MaturityMonthYear = msg.GetString(TagMaturityMonthYear)
	m.Body.MaturityDate = msg.GetString(TagMaturityDate)
	if cr, ok := msg.GetFloat(TagCouponRate); ok {
		m.Body.CouponRate = cr
	}
	m.Body.SecurityDesc = msg.GetString(TagSecurityDesc)
	m.Body.Side = msg.GetString(TagSide)
	m.Body.TransactTime = msg.GetString(TagTransactTime)
	m.Body.HandlInst = msg.GetString(TagHandlInst)
	m.Body.OrdType = msg.GetString(TagOrdType)
	if px, ok := msg.GetFloat(TagPrice); ok {
		m.Body.Price = px
	}
	if sp, ok := msg.GetFloat(TagStopPx); ok {
		m.Body.StopPx = sp
	}
	m.Body.Currency = msg.GetString(TagCurrency)
	m.Body.TimeInForce = msg.GetString(TagTimeInForce)
	if qty, ok := msg.GetFloat(TagQuantity); ok {
		m.Body.OrderQty = qty
	}
	if qty, ok := msg.GetFloat(TagCashOrderQty); ok {
		m.Body.CashOrderQty = qty
	}
	m.Body.PeggedRef = msg.GetString(TagPeggedRef)
	m.Body.DiscloseFlag = msg.GetString(TagDiscloseFlag)
	if mo, ok := msg.GetBool(TagManualOrder); ok {
		m.Body.ManualOrder = mo
	}

	return nil
}

type ExecutionReportMsg struct {
	Header
	Body    ExecutionReportBody
	Trailer Trailer
}

type ExecutionReportBody struct {
	OrderID          string
	SecondaryOrderID string
	ClOrdID          string
	OrigClOrdID      string
	ClientID         string
	ExecBroker       string
	QuoteReqID       string
	ExecID           string
	ExecRefID        string
	ExecTransType    string
	ExecType         string
	OrdStatus        string
	OrdRejReason     int
	Account          string
	SettlDate        string
	SettlDate2       string
	Symbol           string
	Side             string
	OrderQty         float64
	OrderQty2        float64
	Price            float64
	StopPx           float64
	LastPx           float64
	LastQty          float64
	Currency         string
	TimeInForce      string
	ExecInst         string
	Rule80A          string
	LastMovement     string
	LastCapacity     string
	CumQty           float64
	LeavesQty        float64
	AvgPx            float64
	TradeDate        string
	TransactTime     string
	ReportToExch     bool
	Commission       float64
	CommCurrency     string
	CommissionData   string
	Yield            float64
	YieldType        string
	Text             string
	EncodedTextLen   int
	EncodedText      string
}

func (m *ExecutionReportMsg) MsgType() string {
	return MsgTypeExecutionReport
}

func (m *ExecutionReportMsg) Encode() []byte {
	buf := GetBuffer(512)
	defer PutBuffer(buf)

	buf = appendField(buf, TagBeginString, m.Header.BeginString)
	buf = appendField(buf, TagMsgType, MsgTypeExecutionReport)
	buf = appendField(buf, TagSenderCompID, m.Header.SenderCompID)
	buf = appendField(buf, TagTargetCompID, m.Header.TargetCompID)
	buf = appendField(buf, TagMsgSeqNum, formatInt(int64(m.Header.MsgSeqNum)))
	buf = appendField(buf, TagSendingTime, m.Header.SendingTime)

	buf = appendField(buf, TagOrdID, m.Body.OrderID)
	buf = appendField(buf, TagSecondaryOrderID, m.Body.SecondaryOrderID)
	buf = appendField(buf, TagClOrdID, m.Body.ClOrdID)
	buf = appendField(buf, TagOrigClOrdID, m.Body.OrigClOrdID)
	buf = appendField(buf, TagClientID, m.Body.ClientID)
	buf = appendField(buf, TagExecBroker, m.Body.ExecBroker)
	buf = appendField(buf, TagQuoteReqID, m.Body.QuoteReqID)
	buf = appendField(buf, TagExecID, m.Body.ExecID)
	buf = appendField(buf, TagExecRefID, m.Body.ExecRefID)
	buf = appendField(buf, TagExecTransType, m.Body.ExecTransType)
	buf = appendField(buf, TagExecType, m.Body.ExecType)
	buf = appendField(buf, TagOrdStatus, m.Body.OrdStatus)
	if m.Body.OrdRejReason > 0 {
		buf = appendField(buf, TagOrdRejReason, formatInt(int64(m.Body.OrdRejReason)))
	}
	buf = appendField(buf, TagAccount, m.Body.Account)
	buf = appendField(buf, TagSettlDate, m.Body.SettlDate)
	buf = appendField(buf, TagSettlDate2, m.Body.SettlDate2)
	buf = appendField(buf, TagSymbol, m.Body.Symbol)
	buf = appendField(buf, TagSide, m.Body.Side)
	if m.Body.OrderQty != 0 {
		buf = appendField(buf, TagQuantity, strconv.FormatFloat(m.Body.OrderQty, 'f', -1, 64))
	}
	if m.Body.OrderQty2 != 0 {
		buf = appendField(buf, TagOrderQty2, strconv.FormatFloat(m.Body.OrderQty2, 'f', -1, 64))
	}
	if m.Body.Price != 0 {
		buf = appendField(buf, TagPrice, strconv.FormatFloat(m.Body.Price, 'f', -1, 64))
	}
	if m.Body.StopPx != 0 {
		buf = appendField(buf, TagStopPx, strconv.FormatFloat(m.Body.StopPx, 'f', -1, 64))
	}
	if m.Body.LastPx != 0 {
		buf = appendField(buf, TagLastPx, strconv.FormatFloat(m.Body.LastPx, 'f', -1, 64))
	}
	if m.Body.LastQty != 0 {
		buf = appendField(buf, TagLastQty, strconv.FormatFloat(m.Body.LastQty, 'f', -1, 64))
	}
	buf = appendField(buf, TagCurrency, m.Body.Currency)
	buf = appendField(buf, TagTimeInForce, m.Body.TimeInForce)
	buf = appendField(buf, TagExecInst, m.Body.ExecInst)
	buf = appendField(buf, TagRule80A, m.Body.Rule80A)
	buf = appendField(buf, TagLastMovement, m.Body.LastMovement)
	buf = appendField(buf, TagLastCapacity, m.Body.LastCapacity)
	if m.Body.CumQty != 0 {
		buf = appendField(buf, TagCumQty, strconv.FormatFloat(m.Body.CumQty, 'f', -1, 64))
	}
	if m.Body.LeavesQty != 0 {
		buf = appendField(buf, TagLeavesQty, strconv.FormatFloat(m.Body.LeavesQty, 'f', -1, 64))
	}
	if m.Body.AvgPx != 0 {
		buf = appendField(buf, TagAvgPx, strconv.FormatFloat(m.Body.AvgPx, 'f', -1, 64))
	}
	buf = appendField(buf, TagTradeDate, m.Body.TradeDate)
	buf = appendField(buf, TagTransactTime, m.Body.TransactTime)
	if m.Body.ReportToExch {
		buf = appendField(buf, TagReportToExch, "Y")
	}
	if m.Body.Commission != 0 {
		buf = appendField(buf, TagCommission, strconv.FormatFloat(m.Body.Commission, 'f', -1, 64))
	}
	buf = appendField(buf, TagCommCurrency, m.Body.CommCurrency)
	buf = appendField(buf, TagCommissionData, m.Body.CommissionData)
	if m.Body.Yield != 0 {
		buf = appendField(buf, TagYield, strconv.FormatFloat(m.Body.Yield, 'f', -1, 64))
	}
	buf = appendField(buf, TagYieldType, m.Body.YieldType)
	buf = appendField(buf, TagText, m.Body.Text)
	if m.Body.EncodedTextLen > 0 {
		buf = appendField(buf, TagEncodedTextLen, formatInt(int64(m.Body.EncodedTextLen)))
	}
	buf = appendField(buf, TagEncodedText, m.Body.EncodedText)

	return serializeWithChecksum(buf)
}

func (m *ExecutionReportMsg) Decode(data []byte) error {
	msg := GetMessage()
	defer PutMessage(msg)

	if err := parseMessage(data, msg); err != nil {
		return err
	}

	m.Header.BeginString = msg.GetString(TagBeginString)
	m.Header.MsgType = msg.GetString(TagMsgType)
	m.Header.SenderCompID = msg.GetString(TagSenderCompID)
	m.Header.TargetCompID = msg.GetString(TagTargetCompID)
	if seq, ok := msg.GetInt(TagMsgSeqNum); ok {
		m.Header.MsgSeqNum = int(seq)
	}
	m.Header.SendingTime = msg.GetString(TagSendingTime)

	m.Body.OrderID = msg.GetString(TagOrdID)
	m.Body.SecondaryOrderID = msg.GetString(TagSecondaryOrderID)
	m.Body.ClOrdID = msg.GetString(TagClOrdID)
	m.Body.OrigClOrdID = msg.GetString(TagOrigClOrdID)
	m.Body.ClientID = msg.GetString(TagClientID)
	m.Body.ExecBroker = msg.GetString(TagExecBroker)
	m.Body.QuoteReqID = msg.GetString(TagQuoteReqID)
	m.Body.ExecID = msg.GetString(TagExecID)
	m.Body.ExecRefID = msg.GetString(TagExecRefID)
	m.Body.ExecTransType = msg.GetString(TagExecTransType)
	m.Body.ExecType = msg.GetString(TagExecType)
	m.Body.OrdStatus = msg.GetString(TagOrdStatus)
	if r, ok := msg.GetInt(TagOrdRejReason); ok {
		m.Body.OrdRejReason = int(r)
	}
	m.Body.Account = msg.GetString(TagAccount)
	m.Body.SettlDate = msg.GetString(TagSettlDate)
	m.Body.SettlDate2 = msg.GetString(TagSettlDate2)
	m.Body.Symbol = msg.GetString(TagSymbol)
	m.Body.Side = msg.GetString(TagSide)
	if qty, ok := msg.GetFloat(TagQuantity); ok {
		m.Body.OrderQty = qty
	}
	if px, ok := msg.GetFloat(TagPrice); ok {
		m.Body.Price = px
	}
	if lpx, ok := msg.GetFloat(TagLastPx); ok {
		m.Body.LastPx = lpx
	}
	if lqty, ok := msg.GetFloat(TagLastQty); ok {
		m.Body.LastQty = lqty
	}
	m.Body.Currency = msg.GetString(TagCurrency)
	m.Body.TransactTime = msg.GetString(TagTransactTime)
	if qty, ok := msg.GetFloat(TagCumQty); ok {
		m.Body.CumQty = qty
	}
	if qty, ok := msg.GetFloat(TagLeavesQty); ok {
		m.Body.LeavesQty = qty
	}
	if px, ok := msg.GetFloat(TagAvgPx); ok {
		m.Body.AvgPx = px
	}
	m.Body.Text = msg.GetString(TagText)
	m.Body.EncodedText = msg.GetString(TagEncodedText)

	return nil
}

type OrderCancelRequestMsg struct {
	Header
	Body    OrderCancelRequestBody
	Trailer Trailer
}

type OrderCancelRequestBody struct {
	OrderID      string
	ClOrdID      string
	OrigClOrdID  string
	QuoteReqID   string
	Symbol       string
	Side         string
	TransactTime string
	OrderQty     float64
	CashOrderQty float64
}

func (m *OrderCancelRequestMsg) MsgType() string {
	return MsgTypeOrderCancelRequest
}

func (m *OrderCancelRequestMsg) Encode() []byte {
	buf := GetBuffer(256)
	defer PutBuffer(buf)

	buf = appendField(buf, TagBeginString, m.Header.BeginString)
	buf = appendField(buf, TagMsgType, MsgTypeOrderCancelRequest)
	buf = appendField(buf, TagSenderCompID, m.Header.SenderCompID)
	buf = appendField(buf, TagTargetCompID, m.Header.TargetCompID)
	buf = appendField(buf, TagMsgSeqNum, formatInt(int64(m.Header.MsgSeqNum)))
	buf = appendField(buf, TagSendingTime, m.Header.SendingTime)

	buf = appendField(buf, TagOrdID, m.Body.OrderID)
	buf = appendField(buf, TagClOrdID, m.Body.ClOrdID)
	buf = appendField(buf, TagOrigClOrdID, m.Body.OrigClOrdID)
	buf = appendField(buf, TagQuoteReqID, m.Body.QuoteReqID)
	buf = appendField(buf, TagSymbol, m.Body.Symbol)
	buf = appendField(buf, TagSide, m.Body.Side)
	buf = appendField(buf, TagTransactTime, m.Body.TransactTime)
	if m.Body.OrderQty != 0 {
		buf = appendField(buf, TagQuantity, strconv.FormatFloat(m.Body.OrderQty, 'f', -1, 64))
	}
	if m.Body.CashOrderQty != 0 {
		buf = appendField(buf, TagCashOrderQty, strconv.FormatFloat(m.Body.CashOrderQty, 'f', -1, 64))
	}

	return serializeWithChecksum(buf)
}

func (m *OrderCancelRequestMsg) Decode(data []byte) error {
	msg := GetMessage()
	defer PutMessage(msg)

	if err := parseMessage(data, msg); err != nil {
		return err
	}

	m.Header.BeginString = msg.GetString(TagBeginString)
	m.Header.MsgType = msg.GetString(TagMsgType)
	m.Header.SenderCompID = msg.GetString(TagSenderCompID)
	m.Header.TargetCompID = msg.GetString(TagTargetCompID)
	m.Header.SendingTime = msg.GetString(TagSendingTime)

	m.Body.OrderID = msg.GetString(TagOrdID)
	m.Body.ClOrdID = msg.GetString(TagClOrdID)
	m.Body.OrigClOrdID = msg.GetString(TagOrigClOrdID)
	m.Body.QuoteReqID = msg.GetString(TagQuoteReqID)
	m.Body.Symbol = msg.GetString(TagSymbol)
	m.Body.Side = msg.GetString(TagSide)
	m.Body.TransactTime = msg.GetString(TagTransactTime)
	if qty, ok := msg.GetFloat(TagQuantity); ok {
		m.Body.OrderQty = qty
	}
	if qty, ok := msg.GetFloat(TagCashOrderQty); ok {
		m.Body.CashOrderQty = qty
	}

	return nil
}

type OrderCancelReplaceMsg struct {
	Header
	Body    OrderCancelReplaceBody
	Trailer Trailer
}

type OrderCancelReplaceBody struct {
	OrderID      string
	ClOrdID      string
	OrigClOrdID  string
	QuoteReqID   string
	Symbol       string
	Side         string
	TransactTime string
	OrderQty     float64
	Price        float64
	StopPx       float64
	OrdType      string
	Side2        string
	OrderQty2    float64
	Currency     string
	TimeInForce  string
	HandlInst    string
	ExecInst     string
}

func (m *OrderCancelReplaceMsg) MsgType() string {
	return MsgTypeOrderCancelReplace
}

func (m *OrderCancelReplaceMsg) Encode() []byte {
	buf := GetBuffer(256)
	defer PutBuffer(buf)

	buf = appendField(buf, TagBeginString, m.Header.BeginString)
	buf = appendField(buf, TagMsgType, MsgTypeOrderCancelReplace)
	buf = appendField(buf, TagSenderCompID, m.Header.SenderCompID)
	buf = appendField(buf, TagTargetCompID, m.Header.TargetCompID)
	buf = appendField(buf, TagMsgSeqNum, formatInt(int64(m.Header.MsgSeqNum)))
	buf = appendField(buf, TagSendingTime, m.Header.SendingTime)

	buf = appendField(buf, TagOrdID, m.Body.OrderID)
	buf = appendField(buf, TagClOrdID, m.Body.ClOrdID)
	buf = appendField(buf, TagOrigClOrdID, m.Body.OrigClOrdID)
	buf = appendField(buf, TagQuoteReqID, m.Body.QuoteReqID)
	buf = appendField(buf, TagSymbol, m.Body.Symbol)
	buf = appendField(buf, TagSide, m.Body.Side)
	buf = appendField(buf, TagTransactTime, m.Body.TransactTime)
	if m.Body.OrderQty != 0 {
		buf = appendField(buf, TagQuantity, strconv.FormatFloat(m.Body.OrderQty, 'f', -1, 64))
	}
	if m.Body.Price != 0 {
		buf = appendField(buf, TagPrice, strconv.FormatFloat(m.Body.Price, 'f', -1, 64))
	}
	if m.Body.StopPx != 0 {
		buf = appendField(buf, TagStopPx, strconv.FormatFloat(m.Body.StopPx, 'f', -1, 64))
	}
	buf = appendField(buf, TagOrdType, m.Body.OrdType)
	buf = appendField(buf, TagSide2, m.Body.Side2)
	if m.Body.OrderQty2 != 0 {
		buf = appendField(buf, TagOrderQty2, strconv.FormatFloat(m.Body.OrderQty2, 'f', -1, 64))
	}
	buf = appendField(buf, TagCurrency, m.Body.Currency)
	buf = appendField(buf, TagTimeInForce, m.Body.TimeInForce)
	buf = appendField(buf, TagHandlInst, m.Body.HandlInst)
	buf = appendField(buf, TagExecInst, m.Body.ExecInst)

	return serializeWithChecksum(buf)
}

func (m *OrderCancelReplaceMsg) Decode(data []byte) error {
	msg := GetMessage()
	defer PutMessage(msg)

	if err := parseMessage(data, msg); err != nil {
		return err
	}

	m.Header.BeginString = msg.GetString(TagBeginString)
	m.Header.MsgType = msg.GetString(TagMsgType)
	m.Header.SenderCompID = msg.GetString(TagSenderCompID)
	m.Header.TargetCompID = msg.GetString(TagTargetCompID)
	m.Header.SendingTime = msg.GetString(TagSendingTime)

	m.Body.OrderID = msg.GetString(TagOrdID)
	m.Body.ClOrdID = msg.GetString(TagClOrdID)
	m.Body.OrigClOrdID = msg.GetString(TagOrigClOrdID)
	m.Body.Symbol = msg.GetString(TagSymbol)
	m.Body.Side = msg.GetString(TagSide)
	m.Body.TransactTime = msg.GetString(TagTransactTime)
	if qty, ok := msg.GetFloat(TagQuantity); ok {
		m.Body.OrderQty = qty
	}
	if px, ok := msg.GetFloat(TagPrice); ok {
		m.Body.Price = px
	}
	m.Body.OrdType = msg.GetString(TagOrdType)
	m.Body.Currency = msg.GetString(TagCurrency)
	m.Body.TimeInForce = msg.GetString(TagTimeInForce)

	return nil
}

type OrderStatusRequestMsg struct {
	Header
	Body    OrderStatusRequestBody
	Trailer Trailer
}

type OrderStatusRequestBody struct {
	ClOrdID    string
	OrderID    string
	QuoteReqID string
	Symbol     string
	Side       string
}

func (m *OrderStatusRequestMsg) MsgType() string {
	return MsgTypeOrderStatusRequest
}

func (m *OrderStatusRequestMsg) Encode() []byte {
	buf := GetBuffer(128)
	defer PutBuffer(buf)

	buf = appendField(buf, TagBeginString, m.Header.BeginString)
	buf = appendField(buf, TagMsgType, MsgTypeOrderStatusRequest)
	buf = appendField(buf, TagSenderCompID, m.Header.SenderCompID)
	buf = appendField(buf, TagTargetCompID, m.Header.TargetCompID)
	buf = appendField(buf, TagMsgSeqNum, formatInt(int64(m.Header.MsgSeqNum)))
	buf = appendField(buf, TagSendingTime, m.Header.SendingTime)

	buf = appendField(buf, TagClOrdID, m.Body.ClOrdID)
	buf = appendField(buf, TagOrdID, m.Body.OrderID)
	buf = appendField(buf, TagQuoteReqID, m.Body.QuoteReqID)
	buf = appendField(buf, TagSymbol, m.Body.Symbol)
	buf = appendField(buf, TagSide, m.Body.Side)

	return serializeWithChecksum(buf)
}

func (m *OrderStatusRequestMsg) Decode(data []byte) error {
	msg := GetMessage()
	defer PutMessage(msg)

	if err := parseMessage(data, msg); err != nil {
		return err
	}

	m.Header.BeginString = msg.GetString(TagBeginString)
	m.Header.MsgType = msg.GetString(TagMsgType)
	m.Header.SenderCompID = msg.GetString(TagSenderCompID)
	m.Header.TargetCompID = msg.GetString(TagTargetCompID)
	m.Header.SendingTime = msg.GetString(TagSendingTime)

	m.Body.ClOrdID = msg.GetString(TagClOrdID)
	m.Body.OrderID = msg.GetString(TagOrdID)
	m.Body.QuoteReqID = msg.GetString(TagQuoteReqID)
	m.Body.Symbol = msg.GetString(TagSymbol)
	m.Body.Side = msg.GetString(TagSide)

	return nil
}

type OrderCancelRejectMsg struct {
	Header
	Body    OrderCancelRejectBody
	Trailer Trailer
}

type OrderCancelRejectBody struct {
	OrderID          string
	ClOrdID          string
	OrigClOrdID      string
	OrdStatus        string
	OrdRejReason     int
	Text             string
	CxlRejResponseTo string
}

func (m *OrderCancelRejectMsg) MsgType() string {
	return MsgTypeOrderCancelReject
}

func (m *OrderCancelRejectMsg) Encode() []byte {
	buf := GetBuffer(128)
	defer PutBuffer(buf)

	buf = appendField(buf, TagBeginString, m.Header.BeginString)
	buf = appendField(buf, TagMsgType, MsgTypeOrderCancelReject)
	buf = appendField(buf, TagSenderCompID, m.Header.SenderCompID)
	buf = appendField(buf, TagTargetCompID, m.Header.TargetCompID)
	buf = appendField(buf, TagMsgSeqNum, formatInt(int64(m.Header.MsgSeqNum)))
	buf = appendField(buf, TagSendingTime, m.Header.SendingTime)

	buf = appendField(buf, TagOrdID, m.Body.OrderID)
	buf = appendField(buf, TagClOrdID, m.Body.ClOrdID)
	buf = appendField(buf, TagOrigClOrdID, m.Body.OrigClOrdID)
	buf = appendField(buf, TagOrdStatus, m.Body.OrdStatus)
	if m.Body.OrdRejReason > 0 {
		buf = appendField(buf, TagOrdRejReason, formatInt(int64(m.Body.OrdRejReason)))
	}
	buf = appendField(buf, TagText, m.Body.Text)
	buf = appendField(buf, TagCxlRejResponseTo, m.Body.CxlRejResponseTo)

	return serializeWithChecksum(buf)
}

func (m *OrderCancelRejectMsg) Decode(data []byte) error {
	msg := GetMessage()
	defer PutMessage(msg)

	if err := parseMessage(data, msg); err != nil {
		return err
	}

	m.Header.BeginString = msg.GetString(TagBeginString)
	m.Header.MsgType = msg.GetString(TagMsgType)
	m.Header.SenderCompID = msg.GetString(TagSenderCompID)
	m.Header.TargetCompID = msg.GetString(TagTargetCompID)
	m.Header.SendingTime = msg.GetString(TagSendingTime)

	m.Body.OrderID = msg.GetString(TagOrdID)
	m.Body.ClOrdID = msg.GetString(TagClOrdID)
	m.Body.OrigClOrdID = msg.GetString(TagOrigClOrdID)
	m.Body.OrdStatus = msg.GetString(TagOrdStatus)
	if r, ok := msg.GetInt(TagOrdRejReason); ok {
		m.Body.OrdRejReason = int(r)
	}
	m.Body.Text = msg.GetString(TagText)
	m.Body.CxlRejResponseTo = msg.GetString(TagCxlRejResponseTo)

	return nil
}

type HeartbeatMsg struct {
	Header
	Body    HeartbeatBody
	Trailer Trailer
}

type HeartbeatBody struct {
	TestReqID string
}

func (m *HeartbeatMsg) MsgType() string {
	return MsgTypeHeartbeat
}

func (m *HeartbeatMsg) Encode() []byte {
	buf := GetBuffer(64)
	defer PutBuffer(buf)

	buf = appendField(buf, TagBeginString, m.Header.BeginString)
	buf = appendField(buf, TagMsgType, MsgTypeHeartbeat)
	buf = appendField(buf, TagSenderCompID, m.Header.SenderCompID)
	buf = appendField(buf, TagTargetCompID, m.Header.TargetCompID)
	buf = appendField(buf, TagMsgSeqNum, formatInt(int64(m.Header.MsgSeqNum)))
	buf = appendField(buf, TagSendingTime, m.Header.SendingTime)

	if m.Body.TestReqID != "" {
		buf = appendField(buf, TagTestReqID, m.Body.TestReqID)
	}

	return serializeWithChecksum(buf)
}

func (m *HeartbeatMsg) Decode(data []byte) error {
	msg := GetMessage()
	defer PutMessage(msg)

	if err := parseMessage(data, msg); err != nil {
		return err
	}

	m.Header.BeginString = msg.GetString(TagBeginString)
	m.Header.MsgType = msg.GetString(TagMsgType)
	m.Header.SenderCompID = msg.GetString(TagSenderCompID)
	m.Header.TargetCompID = msg.GetString(TagTargetCompID)
	m.Header.SendingTime = msg.GetString(TagSendingTime)
	m.Body.TestReqID = msg.GetString(TagTestReqID)

	return nil
}

type TestRequestMsg struct {
	Header
	Body    TestRequestBody
	Trailer Trailer
}

type TestRequestBody struct {
	TestReqID string
}

func (m *TestRequestMsg) MsgType() string {
	return MsgTypeTestRequest
}

func (m *TestRequestMsg) Encode() []byte {
	buf := GetBuffer(64)
	defer PutBuffer(buf)

	buf = appendField(buf, TagBeginString, m.Header.BeginString)
	buf = appendField(buf, TagMsgType, MsgTypeTestRequest)
	buf = appendField(buf, TagSenderCompID, m.Header.SenderCompID)
	buf = appendField(buf, TagTargetCompID, m.Header.TargetCompID)
	buf = appendField(buf, TagMsgSeqNum, formatInt(int64(m.Header.MsgSeqNum)))
	buf = appendField(buf, TagSendingTime, m.Header.SendingTime)

	buf = appendField(buf, TagTestReqID, m.Body.TestReqID)

	return serializeWithChecksum(buf)
}

func (m *TestRequestMsg) Decode(data []byte) error {
	msg := GetMessage()
	defer PutMessage(msg)

	if err := parseMessage(data, msg); err != nil {
		return err
	}

	m.Header.BeginString = msg.GetString(TagBeginString)
	m.Header.MsgType = msg.GetString(TagMsgType)
	m.Header.SenderCompID = msg.GetString(TagSenderCompID)
	m.Header.TargetCompID = msg.GetString(TagTargetCompID)
	m.Header.SendingTime = msg.GetString(TagSendingTime)
	m.Body.TestReqID = msg.GetString(TagTestReqID)

	return nil
}

type ResendRequestMsg struct {
	Header
	Body    ResendRequestBody
	Trailer Trailer
}

type ResendRequestBody struct {
	BeginSeqNo int
	EndSeqNo   int
}

func (m *ResendRequestMsg) MsgType() string {
	return MsgTypeResendRequest
}

func (m *ResendRequestMsg) Encode() []byte {
	buf := GetBuffer(64)
	defer PutBuffer(buf)

	buf = appendField(buf, TagBeginString, m.Header.BeginString)
	buf = appendField(buf, TagMsgType, MsgTypeResendRequest)
	buf = appendField(buf, TagSenderCompID, m.Header.SenderCompID)
	buf = appendField(buf, TagTargetCompID, m.Header.TargetCompID)
	buf = appendField(buf, TagMsgSeqNum, formatInt(int64(m.Header.MsgSeqNum)))
	buf = appendField(buf, TagSendingTime, m.Header.SendingTime)

	buf = appendField(buf, TagBeginSeqNo, formatInt(int64(m.Body.BeginSeqNo)))
	buf = appendField(buf, TagEndSeqNo, formatInt(int64(m.Body.EndSeqNo)))

	return serializeWithChecksum(buf)
}

func (m *ResendRequestMsg) Decode(data []byte) error {
	msg := GetMessage()
	defer PutMessage(msg)

	if err := parseMessage(data, msg); err != nil {
		return err
	}

	m.Header.BeginString = msg.GetString(TagBeginString)
	m.Header.MsgType = msg.GetString(TagMsgType)
	m.Header.SenderCompID = msg.GetString(TagSenderCompID)
	m.Header.TargetCompID = msg.GetString(TagTargetCompID)
	m.Header.SendingTime = msg.GetString(TagSendingTime)

	if seq, ok := msg.GetInt(TagBeginSeqNo); ok {
		m.Body.BeginSeqNo = int(seq)
	}
	if seq, ok := msg.GetInt(TagEndSeqNo); ok {
		m.Body.EndSeqNo = int(seq)
	}

	return nil
}

type RejectMsg struct {
	Header
	Body    RejectBody
	Trailer Trailer
}

type RejectBody struct {
	RefMsgType          string
	SessionRejectReason int
	Text                string
}

func (m *RejectMsg) MsgType() string {
	return MsgTypeReject
}

func (m *RejectMsg) Encode() []byte {
	buf := GetBuffer(64)
	defer PutBuffer(buf)

	buf = appendField(buf, TagBeginString, m.Header.BeginString)
	buf = appendField(buf, TagMsgType, MsgTypeReject)
	buf = appendField(buf, TagSenderCompID, m.Header.SenderCompID)
	buf = appendField(buf, TagTargetCompID, m.Header.TargetCompID)
	buf = appendField(buf, TagMsgSeqNum, formatInt(int64(m.Header.MsgSeqNum)))
	buf = appendField(buf, TagSendingTime, m.Header.SendingTime)

	buf = appendField(buf, TagRefMsgType, m.Body.RefMsgType)
	if m.Body.SessionRejectReason > 0 {
		buf = appendField(buf, TagSessionRejectReason, formatInt(int64(m.Body.SessionRejectReason)))
	}
	buf = appendField(buf, TagText, m.Body.Text)

	return serializeWithChecksum(buf)
}

func (m *RejectMsg) Decode(data []byte) error {
	msg := GetMessage()
	defer PutMessage(msg)

	if err := parseMessage(data, msg); err != nil {
		return err
	}

	m.Header.BeginString = msg.GetString(TagBeginString)
	m.Header.MsgType = msg.GetString(TagMsgType)
	m.Header.SenderCompID = msg.GetString(TagSenderCompID)
	m.Header.TargetCompID = msg.GetString(TagTargetCompID)
	m.Header.SendingTime = msg.GetString(TagSendingTime)

	m.Body.RefMsgType = msg.GetString(TagRefMsgType)
	if r, ok := msg.GetInt(TagSessionRejectReason); ok {
		m.Body.SessionRejectReason = int(r)
	}
	m.Body.Text = msg.GetString(TagText)

	return nil
}

type SequenceResetMsg struct {
	Header
	Body    SequenceResetBody
	Trailer Trailer
}

type SequenceResetBody struct {
	NewSeqNo    int
	GapFillFlag bool
}

func (m *SequenceResetMsg) MsgType() string {
	return MsgTypeSequenceReset
}

func (m *SequenceResetMsg) Encode() []byte {
	buf := GetBuffer(64)
	defer PutBuffer(buf)

	buf = appendField(buf, TagBeginString, m.Header.BeginString)
	buf = appendField(buf, TagMsgType, MsgTypeSequenceReset)
	buf = appendField(buf, TagSenderCompID, m.Header.SenderCompID)
	buf = appendField(buf, TagTargetCompID, m.Header.TargetCompID)
	buf = appendField(buf, TagMsgSeqNum, formatInt(int64(m.Header.MsgSeqNum)))
	buf = appendField(buf, TagSendingTime, m.Header.SendingTime)

	buf = appendField(buf, TagNewSeqNo, formatInt(int64(m.Body.NewSeqNo)))
	if m.Body.GapFillFlag {
		buf = appendField(buf, TagGapFillFlag, "Y")
	}

	return serializeWithChecksum(buf)
}

func (m *SequenceResetMsg) Decode(data []byte) error {
	msg := GetMessage()
	defer PutMessage(msg)

	if err := parseMessage(data, msg); err != nil {
		return err
	}

	m.Header.BeginString = msg.GetString(TagBeginString)
	m.Header.MsgType = msg.GetString(TagMsgType)
	m.Header.SenderCompID = msg.GetString(TagSenderCompID)
	m.Header.TargetCompID = msg.GetString(TagTargetCompID)
	m.Header.SendingTime = msg.GetString(TagSendingTime)

	if seq, ok := msg.GetInt(TagNewSeqNo); ok {
		m.Body.NewSeqNo = int(seq)
	}
	if gap, ok := msg.GetBool(TagGapFillFlag); ok {
		m.Body.GapFillFlag = gap
	}

	return nil
}

type LogoutMsg struct {
	Header
	Body    LogoutBody
	Trailer Trailer
}

type LogoutBody struct {
	Text string
}

func (m *LogoutMsg) MsgType() string {
	return MsgTypeLogout
}

func (m *LogoutMsg) Encode() []byte {
	buf := GetBuffer(64)
	defer PutBuffer(buf)

	buf = appendField(buf, TagBeginString, m.Header.BeginString)
	buf = appendField(buf, TagMsgType, MsgTypeLogout)
	buf = appendField(buf, TagSenderCompID, m.Header.SenderCompID)
	buf = appendField(buf, TagTargetCompID, m.Header.TargetCompID)
	buf = appendField(buf, TagMsgSeqNum, formatInt(int64(m.Header.MsgSeqNum)))
	buf = appendField(buf, TagSendingTime, m.Header.SendingTime)

	if m.Body.Text != "" {
		buf = appendField(buf, TagText, m.Body.Text)
	}

	return serializeWithChecksum(buf)
}

func (m *LogoutMsg) Decode(data []byte) error {
	msg := GetMessage()
	defer PutMessage(msg)

	if err := parseMessage(data, msg); err != nil {
		return err
	}

	m.Header.BeginString = msg.GetString(TagBeginString)
	m.Header.MsgType = msg.GetString(TagMsgType)
	m.Header.SenderCompID = msg.GetString(TagSenderCompID)
	m.Header.TargetCompID = msg.GetString(TagTargetCompID)
	m.Header.SendingTime = msg.GetString(TagSendingTime)
	m.Body.Text = msg.GetString(TagText)

	return nil
}

type LogonMsg struct {
	Header
	Body    LogonBody
	Trailer Trailer
}

type LogonBody struct {
	EncryptMethod           int
	HeartBtInt              int
	ResetSeqNumFlag         bool
	Username                string
	Password                string
	RawDataLength           int
	RawData                 string
	MaxMessageSize          int
	HeartBtGnrtnMode        int
	DefaultApplExtID        int
	EncryptedPasswordMethod int
	EncryptedPassword       string
	EncryptedUsername       string
}

func (m *LogonMsg) MsgType() string {
	return MsgTypeLogon
}

func (m *LogonMsg) Encode() []byte {
	buf := GetBuffer(256)
	defer PutBuffer(buf)

	buf = appendField(buf, TagBeginString, m.Header.BeginString)
	buf = appendField(buf, TagMsgType, MsgTypeLogon)
	buf = appendField(buf, TagSenderCompID, m.Header.SenderCompID)
	buf = appendField(buf, TagTargetCompID, m.Header.TargetCompID)
	buf = appendField(buf, TagMsgSeqNum, formatInt(int64(m.Header.MsgSeqNum)))
	buf = appendField(buf, TagSendingTime, m.Header.SendingTime)

	if m.Body.EncryptMethod > 0 {
		buf = appendField(buf, TagEncryptMethod, formatInt(int64(m.Body.EncryptMethod)))
	}
	if m.Body.HeartBtInt > 0 {
		buf = appendField(buf, TagHeartBtInt, formatInt(int64(m.Body.HeartBtInt)))
	}
	if m.Body.ResetSeqNumFlag {
		buf = appendField(buf, TagResetSeqNumFlag, "Y")
	}
	if m.Body.Username != "" {
		buf = appendField(buf, TagUsername, m.Body.Username)
	}
	if m.Body.Password != "" {
		buf = appendField(buf, TagPassword, m.Body.Password)
	}
	if m.Body.RawDataLength > 0 {
		buf = appendField(buf, TagRawDataLength, formatInt(int64(m.Body.RawDataLength)))
	}
	if m.Body.RawData != "" {
		buf = appendField(buf, TagRawData, m.Body.RawData)
	}
	if m.Body.MaxMessageSize > 0 {
		buf = appendField(buf, TagMaxMessageSize, formatInt(int64(m.Body.MaxMessageSize)))
	}
	if m.Body.HeartBtGnrtnMode > 0 {
		buf = appendField(buf, TagHeartBtGnrtnMode, formatInt(int64(m.Body.HeartBtGnrtnMode)))
	}
	if m.Body.DefaultApplExtID > 0 {
		buf = appendField(buf, TagDefaultApplExtID, formatInt(int64(m.Body.DefaultApplExtID)))
	}
	if m.Body.EncryptedPasswordMethod > 0 {
		buf = appendField(buf, TagEncryptedPasswordMethod, formatInt(int64(m.Body.EncryptedPasswordMethod)))
	}
	if m.Body.EncryptedPassword != "" {
		buf = appendField(buf, TagEncryptedPassword, m.Body.EncryptedPassword)
	}
	if m.Body.EncryptedUsername != "" {
		buf = appendField(buf, TagEncryptedUsername, m.Body.EncryptedUsername)
	}

	return serializeWithChecksum(buf)
}

func (m *LogonMsg) Decode(data []byte) error {
	msg := GetMessage()
	defer PutMessage(msg)

	if err := parseMessage(data, msg); err != nil {
		return err
	}

	m.Header.BeginString = msg.GetString(TagBeginString)
	m.Header.MsgType = msg.GetString(TagMsgType)
	m.Header.SenderCompID = msg.GetString(TagSenderCompID)
	m.Header.TargetCompID = msg.GetString(TagTargetCompID)
	m.Header.SendingTime = msg.GetString(TagSendingTime)

	if e, ok := msg.GetInt(TagEncryptMethod); ok {
		m.Body.EncryptMethod = int(e)
	}
	if h, ok := msg.GetInt(TagHeartBtInt); ok {
		m.Body.HeartBtInt = int(h)
	}
	if r, ok := msg.GetBool(TagResetSeqNumFlag); ok {
		m.Body.ResetSeqNumFlag = r
	}
	m.Body.Username = msg.GetString(TagUsername)
	m.Body.Password = msg.GetString(TagPassword)
	if r, ok := msg.GetInt(TagRawDataLength); ok {
		m.Body.RawDataLength = int(r)
	}
	m.Body.RawData = msg.GetString(TagRawData)
	if mx, ok := msg.GetInt(TagMaxMessageSize); ok {
		m.Body.MaxMessageSize = int(mx)
	}
	if h, ok := msg.GetInt(TagHeartBtGnrtnMode); ok {
		m.Body.HeartBtGnrtnMode = int(h)
	}
	if d, ok := msg.GetInt(TagDefaultApplExtID); ok {
		m.Body.DefaultApplExtID = int(d)
	}
	m.Body.EncryptedPassword = msg.GetString(TagEncryptedPassword)
	m.Body.EncryptedUsername = msg.GetString(TagEncryptedUsername)

	return nil
}

func init() {
	GlobalMsgTypeRegistry.Register(MsgTypeNewOrderSingle, func() MsgType { return &NewOrderSingleMsg{} })
	GlobalMsgTypeRegistry.Register(MsgTypeExecutionReport, func() MsgType { return &ExecutionReportMsg{} })
	GlobalMsgTypeRegistry.Register(MsgTypeOrderCancelRequest, func() MsgType { return &OrderCancelRequestMsg{} })
	GlobalMsgTypeRegistry.Register(MsgTypeOrderCancelReplace, func() MsgType { return &OrderCancelReplaceMsg{} })
	GlobalMsgTypeRegistry.Register(MsgTypeOrderStatusRequest, func() MsgType { return &OrderStatusRequestMsg{} })
	GlobalMsgTypeRegistry.Register(MsgTypeOrderCancelReject, func() MsgType { return &OrderCancelRejectMsg{} })
	GlobalMsgTypeRegistry.Register(MsgTypeHeartbeat, func() MsgType { return &HeartbeatMsg{} })
	GlobalMsgTypeRegistry.Register(MsgTypeTestRequest, func() MsgType { return &TestRequestMsg{} })
	GlobalMsgTypeRegistry.Register(MsgTypeResendRequest, func() MsgType { return &ResendRequestMsg{} })
	GlobalMsgTypeRegistry.Register(MsgTypeReject, func() MsgType { return &RejectMsg{} })
	GlobalMsgTypeRegistry.Register(MsgTypeSequenceReset, func() MsgType { return &SequenceResetMsg{} })
	GlobalMsgTypeRegistry.Register(MsgTypeLogout, func() MsgType { return &LogoutMsg{} })
	GlobalMsgTypeRegistry.Register(MsgTypeLogon, func() MsgType { return &LogonMsg{} })
}

func appendField(buf []byte, tag Tag, value string) []byte {
	buf = strconv.AppendUint(buf, uint64(tag), 10)
	buf = append(buf, '=')
	buf = append(buf, value...)
	buf = append(buf, SOH)
	return buf
}

func serializeWithChecksum(buf []byte) []byte {
	checksum := calculateChecksum(buf)
	buf = appendField(buf, TagCheckSum, checksum)

	bodyStart := findAfterTagValue(buf, TagMsgType)
	if bodyStart > 0 {
		checksumIdx := findTagIndex(buf, TagCheckSum)
		if checksumIdx > 0 {
			bodyLen := checksumIdx - bodyStart
			bodyLenStr := formatInt(int64(bodyLen))
			bodyLenLen := len(bodyLenStr)

			newLen := bodyStart + 2 + bodyLenLen + 1 + (checksumIdx - bodyStart)
			result := make([]byte, newLen+4)

			copy(result[:bodyStart], buf[:bodyStart])
			result[bodyStart] = '9'
			result[bodyStart+1] = '='
			copy(result[bodyStart+2:], bodyLenStr)
			result[bodyStart+2+bodyLenLen] = SOH
			copy(result[bodyStart+2+bodyLenLen+1:], buf[bodyStart:checksumIdx])
			copy(result[bodyStart+2+bodyLenLen+1+(checksumIdx-bodyStart):], buf[checksumIdx:])
			return result
		}
	}
	return buf
}

func findAfterTagValue(buf []byte, tag Tag) int {
	tagStr := strconv.FormatUint(uint64(tag), 10)
	for i := 0; i < len(buf)-len(tagStr)-2; i++ {
		if buf[i] == tagStr[0] && i+len(tagStr)+1 < len(buf) {
			found := true
			for j := 0; j < len(tagStr); j++ {
				if buf[i+j] != tagStr[j] {
					found = false
					break
				}
			}
			if found && buf[i+len(tagStr)] == '=' {
				for j := i + len(tagStr) + 1; j < len(buf); j++ {
					if buf[j] == SOH {
						return j + 1
					}
				}
			}
		}
	}
	return -1
}

func findTagIndex(buf []byte, tag Tag) int {
	tagStr := strconv.FormatUint(uint64(tag), 10)
	for i := 0; i < len(buf)-len(tagStr)-2; i++ {
		if buf[i] == tagStr[0] && i+len(tagStr)+1 < len(buf) {
			found := true
			for j := 0; j < len(tagStr); j++ {
				if buf[i+j] != tagStr[j] {
					found = false
					break
				}
			}
			if found && buf[i+len(tagStr)] == '=' {
				return i
			}
		}
	}
	return -1
}

func calculateChecksum(data []byte) string {
	var cs uint8
	for i := 0; i < len(data); i++ {
		cs ^= data[i]
	}
	return string([]byte{
		hexDigit(cs >> 4),
		hexDigit(cs & 0x0F),
		'0',
	})
}

func parseMessage(data []byte, msg *Message) error {
	fields := extractFields(data)
	for _, f := range fields {
		if len(f.tag) > 0 && len(f.value) > 0 {
			tag := parseTag(f.tag)
			if tag == TagBeginString {
				msg.BeginString = f.value
			} else if tag == TagMsgType {
				msg.MsgType = f.value
			} else {
				msg.AddField(tag, f.value)
			}
		}
	}
	return nil
}

type fieldPair struct {
	tag   string
	value string
}

func extractFields(data []byte) []fieldPair {
	var fields []fieldPair
	var current fieldPair
	inTag := true

	for i := 0; i < len(data); i++ {
		if data[i] == '=' {
			inTag = false
		} else if data[i] == SOH {
			if len(current.tag) > 0 {
				fields = append(fields, current)
			}
			current = fieldPair{}
			inTag = true
		} else if inTag {
			current.tag += string(data[i])
		} else {
			current.value += string(data[i])
		}
	}

	return fields
}

func parseTag(tagStr string) Tag {
	tag, _ := strconv.ParseUint(tagStr, 10, 16)
	return Tag(tag)
}
