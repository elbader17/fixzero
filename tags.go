package fixzero

// Tag represents a FIX field tag number.
// Tags are typically 1-9999.
type Tag uint16

// Standard FIX field tags.
const (
	TagUnknown         Tag = 0
	TagBeginString     Tag = 8  // FIX version (e.g., "FIX.4.4")
	TagBodyLength      Tag = 9  // Message length
	TagMsgType         Tag = 35 // Message type
	TagSenderCompID    Tag = 49 // Sender CompID
	TagTargetCompID    Tag = 56 // Target CompID
	TagMsgSeqNum       Tag = 34 // Sequence number
	TagSendingTime     Tag = 52 // Sending time
	TagOrigSendingTime Tag = 52 // Orig Sending time
	TagPossDupFlag     Tag = 43 // PossDupFlag

	// Order related
	TagOrdID   Tag = 37 // OrderID
	TagClOrdID Tag = 11 // ClOrdID
	TagOrderID Tag = 37 // OrderID

	// Instrument
	TagSymbol           Tag = 55  // Symbol
	TagSecurityID       Tag = 48  // SecurityID
	TagSecurityIDSource Tag = 22  // SecurityIDSource
	TagSecurityAltID    Tag = 455 // SecurityAltID
	TagProduct          Tag = 460 // Product
	TagCFICode          Tag = 461 // CFICode

	// Side and Type
	TagSide        Tag = 54 // Side: 1=Buy, 2=Sell
	TagOrdType     Tag = 40 // OrderType: 1=Market, 2=Limit, etc.
	TagHandlInst   Tag = 21 // HandlInst: 1=Manual, 2=Auto, 3=AutoWithManual
	TagTimeInForce Tag = 59 // TimeInForce: 0=Day, 1=IOC, 3=GTD

	// Quantities and Prices
	TagQuantity     Tag = 38  // OrderQty
	TagCashOrderQty Tag = 152 // CashOrderQty
	TagLeavesQty    Tag = 151 // LeavesQty
	TagCumQty       Tag = 14  // CumQty
	TagAvgPx        Tag = 6   // AvgPx
	TagPrice        Tag = 44  // Price
	TagStopPx       Tag = 99  // StopPx
	TagYield        Tag = 235 // Yield
	TagYieldType    Tag = 236 // YieldType

	// Execution
	TagExecType      Tag = 150 // ExecType
	TagOrdStatus     Tag = 39  // OrdStatus
	TagExecTransType Tag = 20  // ExecTransType: 0=New, 1=Cancel, 2=Correct, 3=Status
	TagLastPx        Tag = 31  // LastPx
	TagLastQty       Tag = 32  // LastQty
	TagTradeDate     Tag = 75  // TradeDate
	TagTransactTime  Tag = 60  // TransactTime
	TagSettlDate     Tag = 64  // SettlDate
	TagSettlDate2    Tag = 193 // SettlDate2
	TagOrderQty2     Tag = 192 // OrderQty2

	// Currency
	TagCurrency      Tag = 15  // Currency
	TagSettlCurrency Tag = 120 // SettlCurrency

	// Status and Reject
	TagText           Tag = 58  // Text
	TagEncodedTextLen Tag = 354 // EncodedTextLen
	TagEncodedText    Tag = 355 // EncodedText

	// Header fields
	TagDeliverToCompID Tag = 128 // DeliverToCompID
	TagSecureDataLen   Tag = 90  // SecureDataLen
	TagSecureData      Tag = 91  // SecureData
	TagSignatureLen    Tag = 93  // SignatureLen
	TagSignature       Tag = 89  // Signature
	TagCheckSum        Tag = 10  // Checksum (must be last field)

	// Quote
	TagQuoteID    Tag = 117 // QuoteID
	TagQuoteReqID Tag = 131 // QuoteReqID
	TagQuoteType  Tag = 537 // QuoteType
	TagBidPx      Tag = 132 // BidPx
	TagOfferPx    Tag = 133 // OfferPx
	TagBidSize    Tag = 134 // BidSize
	TagOfferSize  Tag = 135 // OfferSize

	// Market Data
	TagMDReqID                 Tag = 262 // MDReqID
	TagSubscriptionRequestType Tag = 263 // SubscriptionRequestType
	TagMarketDepth             Tag = 264 // MarketDepth
	TagMDUpdateType            Tag = 265 // MDUpdateType
	TagMDEntryType             Tag = 269 // MDEntryType
	TagMDEntryPx               Tag = 270 // MDEntryPx
	TagMDEntrySize             Tag = 271 // MDEntrySize
	TagMDEntryDate             Tag = 272 // MDEntryDate
	TagMDEntryTime             Tag = 273 // MDEntryTime
	TagMDEntryPosn             Tag = 290 // MDEntryPosn
	TagMDEntryRefID            Tag = 278 // MDEntryRefID
	TagMDUpdateAction          Tag = 279 // MDUpdateAction

	// Misc
	TagAccount          Tag = 1    // Account
	TagAcctIDSource     Tag = 660  // AcctIDSource
	TagAccountType      Tag = 581  // AccountType
	TagTradeInputSource Tag = 578  // TradeInputSource
	TagTradeInputDevice Tag = 579  // TradeInputDevice
	TagOrderInputDevice Tag = 1002 // OrderInputDevice
	TagVenue            Tag = 1000 // Venue
	TagLocation         Tag = 1001 // Location
	TagGrossTradeAmt    Tag = 381  // GrossTradeAmt
	TagMinBidSize       Tag = 1101 // MinBidSize
	TagMinOfferSize     Tag = 1102 // MinOfferSize
	TagTradeLegRefID    Tag = 1047 // TradeLegRefID
	TagNoLegs           Tag = 555  // NoLegs
	TagLegSymbol        Tag = 600  // LegSymbol
	TagLegSide          Tag = 624  // LegSide
	TagLegRatioQty      Tag = 628  // LegRatioQty
	TagLegPrice         Tag = 680  // LegPrice
	TagLegStipulations  Tag = 688  // LegStipulations

	// Admin
	TagTestReqID               Tag = 112  // TestReqID
	TagRefMsgType              Tag = 372  // RefMsgType
	TagSessionRejectReason     Tag = 373  // SessionRejectReason
	TagGapFillFlag             Tag = 123  // GapFillFlag
	TagNewSeqNo                Tag = 36   // NewSeqNo
	TagOrigClOrdID             Tag = 41   // OrigClOrdID
	TagExecRefID               Tag = 37   // ExecRefID
	TagEncryptMethod           Tag = 98   // EncryptMethod
	TagHeartBtInt              Tag = 108  // HeartBtInt
	TagResetSeqNumFlag         Tag = 141  // ResetSeqNumFlag
	TagUsername                Tag = 553  // Username
	TagPassword                Tag = 554  // Password
	TagRawDataLength           Tag = 95   // RawDataLength
	TagRawData                 Tag = 96   // RawData
	TagMaxMessageSize          Tag = 383  // MaxMessageSize
	TagHeartBtGnrtnMode        Tag = 384  // HeartBtGnrtnMode
	TagDefaultApplExtID        Tag = 1407 // DefaultApplExtID
	TagEncryptedPasswordMethod Tag = 1408 // EncryptedPasswordMethod
	TagEncryptedPassword       Tag = 1409 // EncryptedPassword
	TagEncryptedUsername       Tag = 1410 // EncryptedUsername
	TagEndSeqNo                Tag = 16   // EndSeqNo
	TagBeginSeqNo              Tag = 7    // BeginSeqNo

	// Order extended
	TagSecondaryOrderID  Tag = 198  // SecondaryOrderID
	TagOrdRejReason      Tag = 103  // OrdRejReason
	TagSymbolSfx         Tag = 65   // SymbolSfx
	TagSecurityType      Tag = 167  // SecurityType
	TagMaturityMonthYear Tag = 200  // MaturityMonthYear
	TagMaturityDate      Tag = 541  // MaturityDate
	TagCouponRate        Tag = 223  // CouponRate
	TagSecurityDesc      Tag = 107  // SecurityDesc
	TagRule80A           Tag = 47   // Rule80A
	TagLastMovement      Tag = 205  // LastMovement
	TagLastCapacity      Tag = 29   // LastCapacity
	TagCommission        Tag = 12   // Commission
	TagCommCurrency      Tag = 479  // CommCurrency
	TagCommissionData    Tag = 13   // CommissionData
	TagReportToExch      Tag = 15   // ReportToExch
	TagSide2             Tag = 401  // Side2
	TagExecInst          Tag = 18   // ExecInst
	TagClientID          Tag = 109  // ClientID
	TagExecBroker        Tag = 76   // ExecBroker
	TagManualOrder       Tag = 1002 // ManualOrder
	TagPeggedRef         Tag = 688  // PeggedRef
	TagDiscloseFlag      Tag = 1392 // DiscloseFlag
	TagCxlRejResponseTo  Tag = 434  // CxlRejResponseTo
	TagExecID            Tag = 17   // ExecID
	TagMsgTypeLogon      Tag = 9999 // placeholder for MsgType
)

// MsgType constants for FIX message types.
const (
	// Administrative messages
	MsgTypeHeartbeat             = "0"
	MsgTypeTestRequest           = "1"
	MsgTypeResendRequest         = "2"
	MsgTypeReject                = "3"
	MsgTypeSequenceReset         = "4"
	MsgTypeLogout                = "5"
	MsgTypeLogon                 = "A"
	MsgTypeBusinessMessageReject = "n"

	// Order messages
	MsgTypeNewOrderSingle            = "D"
	MsgTypeOrderCancelRequest        = "F"
	MsgTypeOrderCancelReplace        = "G"
	MsgTypeOrderStatusRequest        = "H"
	MsgTypeOrderCancelReject         = "9"
	MsgTypeTradeCaptureReportRequest = "ad"
	MsgTypeTradeCaptureReport        = "AE"
	MsgTypeTradeCaptureReportAck     = "AT"
	MsgTypeOrderMassCancelRequest    = "q"
	MsgTypeOrderMassCancelReport     = "r"
	MsgTypeOrderMassStatusRequest    = "AF"

	// Execution reports
	MsgTypeExecutionReport = "8"
	MsgTypeTradeReport     = "AE"
	MsgTypeTradeReportAck  = "AT"

	// Quote messages
	MsgTypeQuoteRequest       = "R"
	MsgTypeQuote              = "S"
	MsgTypeQuoteCancel        = "Y"
	MsgTypeQuoteResponse      = "i"
	MsgTypeQuoteStatusRequest = "a"
	MsgTypeQuoteStatusReport  = "AI"
	MsgTypeMassQuote          = "b"

	// Market Data
	MsgTypeMarketDataRequest            = "V"
	MsgTypeMarketDataSnapshot           = "W"
	MsgTypeMarketDataIncrementalRefresh = "X"
	MsgTypeMarketDataRequestReject      = "Y"

	// Position
	MsgTypePositionReport             = "AP"
	MsgTypePositionMaintenanceRequest = "AL"
	MsgTypePositionMaintenanceReport  = "AM"
	MsgTypeRequestForPositions        = "AN"
	MsgTypeRequestForPositionsAck     = "AO"

	// Settlement
	MsgTypeSettlementInstructionRequest = "v"
	MsgTypeSettlementInstruction        = "t"
	MsgTypeSettlementReport             = "AX"
	MsgTypeConfirmation                 = "AK"
	MsgTypeConfirmationRequest          = "BH"

	// Allocation
	MsgTypeAllocationInstruction    = "J"
	MsgTypeAllocationReport         = "AS"
	MsgTypeAllocationInstructionAck = "AQ"
	MsgTypeAllocationReportAck      = "AR"
)

// Side values.
const (
	SideBuy        = "1"
	SideSell       = "2"
	SideSellShort  = "5"
	SideCross      = "8"
	SideCrossShort = "9"
)

// OrdType values.
const (
	OrdTypeMarket        = "1"
	OrdTypeLimit         = "2"
	OrdTypeStop          = "3"
	OrdTypeStopLimit     = "4"
	OrdTypeMarketOnClose = "5"
	OrdTypeLimitOnClose  = "6"
	OrdTypePEG           = "7"
	OrdTypePEG_STOPT     = "8"
	OrdTypePEG_LIMIT     = "9"
)

// TimeInForce values.
const (
	TIFF_Day = "0"
	TIFF_IOC = "1"
	TIFF_GTC = "3"
	TIFF_GTD = "4"
	TIFF_ATC = "5"
	TIFF_GTX = "6"
	TIFF_FOK = "7"
)

// ExecType values.
const (
	ExecTypeNew            = "0"
	ExecTypePartialFill    = "1"
	ExecTypeFilled         = "2"
	ExecTypeDoneForDay     = "3"
	ExecTypeCanceled       = "4"
	ExecTypeReplaced       = "5"
	ExecTypePendingCancel  = "6"
	ExecTypeRejected       = "8"
	ExecTypeSuspended      = "9"
	ExecTypePendingNew     = "A"
	ExecTypeCalculated     = "C"
	ExecTypeExpired        = "E"
	ExecTypeRestated       = "D"
	ExecTypePendingReplace = "F"
)

// OrdStatus values.
const (
	OrdStatusNew            = "0"
	OrdStatusPartialFill    = "1"
	OrdStatusFilled         = "2"
	OrdStatusDoneForDay     = "3"
	OrdStatusCanceled       = "4"
	OrdStatusReplaced       = "5"
	OrdStatusPendingCancel  = "6"
	OrdStatusRejected       = "8"
	OrdStatusSuspended      = "9"
	OrdStatusPendingNew     = "A"
	OrdStatusCalculated     = "C"
	OrdStatusExpired        = "E"
	OrdStatusRestated       = "D"
	OrdStatusPendingReplace = "F"
)
