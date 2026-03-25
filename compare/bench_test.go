package compare

import (
	"bytes"
	"testing"
	
	"github.com/quickfixgo/quickfix"
	
	fixzero "github.com/fixzero/fixzero"
)

// QuickFIX doesn't like our test message format, so we'll create valid ones
func createValidMsgData() []byte {
	msg := quickfix.NewMessage()
	msg.Header.SetString(8, "FIX.4.4")
	msg.Header.SetString(35, "D")
	msg.Body.SetString(11, "ORDER123")
	msg.Body.SetString(55, "AAPL")
	msg.Body.SetString(54, "1")
	msg.Body.SetString(40, "2")
	msg.Body.SetInt(38, 100)
	msg.Body.SetString(44, "150.50")
	return msg.Bytes()
}

var validMsgData = createValidMsgData()

func BenchmarkFixZeroParse(b *testing.B) {
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		msg, err := fixzero.Parse(validMsgData)
		if err != nil {
			b.Fatal(err)
		}
		fixzero.PutMessage(msg)
	}
}

func BenchmarkQuickFIXParse(b *testing.B) {
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		msg := quickfix.NewMessage()
		err := quickfix.ParseMessage(msg, bytes.NewBuffer(validMsgData))
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkQuickFIXNewOrderSingle(b *testing.B) {
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		msg := quickfix.NewMessage()
		msg.Header.SetString(8, "FIX.4.4")
		msg.Header.SetString(35, "D")
		msg.Body.SetString(11, "ORDER123")
		msg.Body.SetString(55, "AAPL")
		msg.Body.SetString(54, "1")
		msg.Body.SetString(40, "2")
		msg.Body.SetInt(38, 100)
		msg.Body.SetString(44, "150.50")
		_ = msg.Bytes()
	}
}

func BenchmarkFixZeroBuilder(b *testing.B) {
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		msg := fixzero.NewBuilder().
			BeginString("FIX.4.4").
			MsgType("D").
			Set(fixzero.TagClOrdID, "ORDER123").
			Set(fixzero.TagSymbol, "AAPL").
			Set(fixzero.TagSide, "1").
			Set(fixzero.TagOrdType, "2").
			SetInt(fixzero.TagQuantity, 100).
			SetFloat(fixzero.TagPrice, 150.50).
			Build()
		
		fixzero.PutMessage(msg)
	}
}

func BenchmarkQuickFIXFieldAccess(b *testing.B) {
	msg := quickfix.NewMessage()
	quickfix.ParseMessage(msg, bytes.NewBuffer(validMsgData))
	
	b.ResetTimer()
	b.ReportAllocs()
	
	var result string
	var err error
	
	for i := 0; i < b.N; i++ {
		result, err = msg.Body.GetString(55)
		if err != nil { b.Fatal(err) }
		result, err = msg.Body.GetString(11)
		if err != nil { b.Fatal(err) }
		result, err = msg.Body.GetString(54)
		if err != nil { b.Fatal(err) }
	}
	_ = result
}

func BenchmarkFixZeroFieldAccess(b *testing.B) {
	msg, _ := fixzero.Parse(validMsgData)
	
	b.ResetTimer()
	b.ReportAllocs()
	
	var result string
	
	for i := 0; i < b.N; i++ {
		result = msg.GetString(fixzero.TagSymbol)
		result = msg.GetString(fixzero.TagClOrdID)
		result = msg.GetString(fixzero.TagSide)
	}
	_ = result
}

func BenchmarkQuickFIXSerialize(b *testing.B) {
	msg := quickfix.NewMessage()
	msg.Header.SetString(8, "FIX.4.4")
	msg.Header.SetString(35, "D")
	msg.Body.SetString(11, "ORDER123")
	msg.Body.SetString(55, "AAPL")
	msg.Body.SetString(54, "1")
	msg.Body.SetString(40, "2")
	msg.Body.SetInt(38, 100)
	msg.Body.SetString(44, "150.50")
	
	b.ReportAllocs()
	b.ResetTimer()
	
	var result []byte
	
	for i := 0; i < b.N; i++ {
		result = msg.Bytes()
	}
	_ = result
}

func BenchmarkFixZeroSerialize(b *testing.B) {
	msg := fixzero.NewBuilder().
		BeginString("FIX.4.4").
		MsgType("D").
		Set(fixzero.TagClOrdID, "ORDER123").
		Set(fixzero.TagSymbol, "AAPL").
		Set(fixzero.TagSide, "1").
		Set(fixzero.TagOrdType, "2").
		SetInt(fixzero.TagQuantity, 100).
		SetFloat(fixzero.TagPrice, 150.50).
		Build()
	
	b.ReportAllocs()
	b.ResetTimer()
	
	var result []byte
	for i := 0; i < b.N; i++ {
		result = fixzero.Serialize(msg)
	}
	_ = result
}
