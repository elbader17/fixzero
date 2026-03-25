package fixzero

import (
	"testing"
)

func BenchmarkParse(b *testing.B) {
	msgData := []byte("8=FIX.4.4\x019=0123\x0135=D\x0134=1\x0149=SENDER\x0156=TARGET\x0152=20240101-10:00:00\x0111=ORDER123\x0155=AAPL\x0154=1\x0140=2\x0138=100\x0144=150.50\x0110=001\x01")
	
	b.ReportAllocs()
	b.ResetTimer()
	
	var msg *Message
	var err error
	
	for i := 0; i < b.N; i++ {
		msg, err = Parse(msgData)
		if err != nil {
			b.Fatal(err)
		}
		PutMessage(msg)
	}
}

func BenchmarkSerialize(b *testing.B) {
	msg := GetMessage()
	msg.BeginString = "FIX.4.4"
	msg.MsgType = "D"
	msg.AddField(TagClOrdID, "ORDER123")
	msg.AddField(TagSymbol, "AAPL")
	msg.AddField(TagSide, "1")
	msg.AddField(TagOrdType, "2")
	msg.AddField(TagQuantity, "100")
	msg.AddField(TagPrice, "150.50")
	
	b.ReportAllocs()
	b.ResetTimer()
	
	var result []byte
	
	for i := 0; i < b.N; i++ {
		result = Serialize(msg)
		_ = result
	}
	
	PutMessage(msg)
}

func BenchmarkFieldAccess(b *testing.B) {
	msgData := []byte("8=FIX.4.4\x019=0123\x0135=D\x0134=1\x0149=SENDER\x0156=TARGET\x0152=20240101-10:00:00\x0111=ORDER123\x0155=AAPL\x0154=1\x0140=2\x0138=100\x0144=150.50\x0110=001\x01")
	
	msg, err := Parse(msgData)
	if err != nil {
		b.Fatal(err)
	}
	
	b.ResetTimer()
	b.ReportAllocs()
	
	var result string
	
	for i := 0; i < b.N; i++ {
		result = msg.GetString(TagSymbol)
		if result == "" {
			b.Fatal("field not found")
		}
		result = msg.GetString(TagClOrdID)
		if result == "" {
			b.Fatal("field not found")
		}
		result = msg.GetString(TagSide)
		if result == "" {
			b.Fatal("field not found")
		}
	}
	
	PutMessage(msg)
	_ = result
}

func BenchmarkBuilder(b *testing.B) {
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		msg := NewBuilder().
			BeginString("FIX.4.4").
			MsgType("D").
			Set(TagClOrdID, "ORDER123").
			Set(TagSymbol, "AAPL").
			Set(TagSide, SideBuy).
			Set(TagOrdType, OrdTypeLimit).
			SetInt(TagQuantity, 100).
			SetFloat(TagPrice, 150.50).
			Build()
		
		PutMessage(msg)
	}
}

func BenchmarkRoundTrip(b *testing.B) {
	msgData := []byte("8=FIX.4.4\x019=0123\x0135=D\x0134=1\x0149=SENDER\x0156=TARGET\x0152=20240101-10:00:00\x0111=ORDER123\x0155=AAPL\x0154=1\x0140=2\x0138=100\x0144=150.50\x0110=001\x01")
	
	b.ReportAllocs()
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		// Parse
		msg, err := Parse(msgData)
		if err != nil {
			b.Fatal(err)
		}
		
		// Serialize
		data := Serialize(msg)
		PutMessage(msg)
		
		// Parse the serialized data
		msg2, err := Parse(data)
		if err != nil {
			b.Fatal(err)
		}
		
		_ = msg2.GetString(TagSymbol)
		PutMessage(msg2)
	}
}
