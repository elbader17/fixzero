package fixzero

import (
	"bytes"
	"testing"
)

func TestParse(t *testing.T) {
	msgData := []byte("8=FIX.4.4\x019=0123\x0135=D\x0134=1\x0149=SENDER\x0156=TARGET\x0152=20240101-10:00:00\x0111=ORDER123\x0155=AAPL\x0154=1\x0140=2\x0138=100\x0144=150.50\x0110=001\x01")
	
	msg, err := Parse(msgData)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	
	if msg.BeginString != "FIX.4.4" {
		t.Errorf("Expected BeginString=FIX.4.4, got %s", msg.BeginString)
	}
	if msg.MsgType != "D" {
		t.Errorf("Expected MsgType=D, got %s", msg.MsgType)
	}
	
	if msg.GetString(TagClOrdID) != "ORDER123" {
		t.Errorf("Expected ClOrdID=ORDER123, got %s", msg.GetString(TagClOrdID))
	}
	if msg.GetString(TagSymbol) != "AAPL" {
		t.Errorf("Expected Symbol=AAPL, got %s", msg.GetString(TagSymbol))
	}
	
	qty, ok := msg.GetInt(TagQuantity)
	if !ok || qty != 100 {
		t.Errorf("Expected Quantity=100, got %d, ok=%v", qty, ok)
	}
	
	price, ok := msg.GetFloat(TagPrice)
	if !ok || price != 150.50 {
		t.Errorf("Expected Price=150.50, got %f, ok=%v", price, ok)
	}
	
	if !msg.Has(TagSymbol) {
		t.Error("Expected Has(Symbol)=true")
	}
	if msg.Has(TagSecurityID) {
		t.Error("Expected Has(SecurityID)=false")
	}
	
	PutMessage(msg)
}

func TestParseReuse(t *testing.T) {
	msgData := []byte("8=FIX.4.4\x019=0123\x0135=D\x0134=1\x0149=SENDER\x0156=TARGET\x0152=20240101-10:00:00\x0111=ORDER123\x0155=AAPL\x0154=1\x0140=2\x0138=100\x0144=150.50\x0110=001\x01")
	
	msg := GetMessage()
	
	err := ParseInto(msgData, msg)
	if err != nil {
		t.Fatalf("ParseInto failed: %v", err)
	}
	
	if msg.GetString(TagSymbol) != "AAPL" {
		t.Errorf("Expected Symbol=AAPL, got %s", msg.GetString(TagSymbol))
	}
	
	err = ParseInto(msgData, msg)
	if err != nil {
		t.Fatalf("ParseInto (reuse) failed: %v", err)
	}
	
	if msg.GetString(TagSymbol) != "AAPL" {
		t.Errorf("Expected Symbol=AAPL after reuse, got %s", msg.GetString(TagSymbol))
	}
	
	PutMessage(msg)
}

func TestSerialize(t *testing.T) {
	msg := GetMessage()
	msg.BeginString = "FIX.4.4"
	msg.MsgType = "D"
	msg.AddField(TagClOrdID, "ORDER123")
	msg.AddField(TagSymbol, "AAPL")
	msg.AddField(TagSide, "1")
	msg.AddField(TagOrdType, "2")
	msg.AddField(TagQuantity, "100")
	msg.AddField(TagPrice, "150.50")
	
	data := Serialize(msg)
	
	if len(data) == 0 {
		t.Error("Serialized data is empty")
	}
	
	// Verify message can be parsed back
	if !bytes.Contains(data, []byte("8=FIX.4.4")) {
		t.Error("Missing BeginString")
	}
	if !bytes.Contains(data, []byte("35=D")) {
		t.Error("Missing MsgType")
	}
	
	t.Logf("Serialized (%d bytes): %q", len(data), string(data))
	
	PutMessage(msg)
}

func TestSerializeParseRoundTrip(t *testing.T) {
	msg := GetMessage()
	msg.BeginString = "FIX.4.4"
	msg.MsgType = "D"
	msg.AddField(TagClOrdID, "ORDER123")
	msg.AddField(TagSymbol, "AAPL")
	msg.AddField(TagSide, "1")
	msg.AddField(TagOrdType, "2")
	msg.AddField(TagQuantity, "100")
	msg.AddField(TagPrice, "150.50")
	
	data := Serialize(msg)
	
	// Parse it back
	msg2, err := Parse(data)
	if err != nil {
		t.Fatalf("Failed to parse serialized message: %v", err)
	}
	
	if msg2.BeginString != "FIX.4.4" {
		t.Errorf("Round-trip failed: expected BeginString=FIX.4.4, got %s", msg2.BeginString)
	}
	if msg2.MsgType != "D" {
		t.Errorf("Round-trip failed: expected MsgType=D, got %s", msg2.MsgType)
	}
	if msg2.GetString(TagSymbol) != "AAPL" {
		t.Errorf("Round-trip failed: expected Symbol=AAPL, got %s", msg2.GetString(TagSymbol))
	}
	
	PutMessage(msg)
	PutMessage(msg2)
}

func TestBuilder(t *testing.T) {
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
	
	if msg.BeginString != "FIX.4.4" {
		t.Errorf("Expected BeginString=FIX.4.4, got %s", msg.BeginString)
	}
	if msg.MsgType != "D" {
		t.Errorf("Expected MsgType=D, got %s", msg.MsgType)
	}
	if msg.GetString(TagSymbol) != "AAPL" {
		t.Errorf("Expected Symbol=AAPL, got %s", msg.GetString(TagSymbol))
	}
	
	qty, ok := msg.GetInt(TagQuantity)
	if !ok || qty != 100 {
		t.Errorf("Expected Quantity=100, got %d", qty)
	}
	
	price, ok := msg.GetFloat(TagPrice)
	if !ok || price != 150.50 {
		t.Errorf("Expected Price=150.50, got %f", price)
	}
	
	PutMessage(msg)
}

func TestBuilderReuse(t *testing.T) {
	b := NewBuilder()
	
	msg1 := b.BeginString("FIX.4.4").MsgType("D").Set(TagSymbol, "AAPL").Build()
	if msg1.GetString(TagSymbol) != "AAPL" {
		t.Error("First build failed")
	}
	PutMessage(msg1)
	
	b.Reset()
	msg2 := b.BeginString("FIX.4.2").MsgType("8").Set(TagSymbol, "GOOG").Build()
	if msg2.GetString(TagSymbol) != "GOOG" {
		t.Error("Second build failed")
	}
	if msg2.BeginString != "FIX.4.2" {
		t.Error("BeginString not reset")
	}
	
	PutMessage(msg2)
}

func TestIterate(t *testing.T) {
	msgData := []byte("8=FIX.4.4\x019=0123\x0135=D\x0134=1\x0149=SENDER\x0156=TARGET\x0152=20240101-10:00:00\x0111=ORDER123\x0155=AAPL\x0154=1\x0140=2\x0138=100\x0144=150.50\x0110=001\x01")
	
	msg, err := Parse(msgData)
	if err != nil {
		t.Fatal(err)
	}
	
	count := 0
	msg.Iterate(func(tag Tag, value string) bool {
		count++
		t.Logf("Field: %d = %s", tag, value)
		return true
	})
	
	if count == 0 {
		t.Error("No fields iterated")
	}
	
	PutMessage(msg)
}

func TestBufferPool(t *testing.T) {
	buf := GetBuffer(1024)
	if cap(buf) < 1024 {
		t.Errorf("Buffer capacity too small: %d", cap(buf))
	}
	
	buf = append(buf, "test data"...)
	
	PutBuffer(buf)
	
	buf2 := GetBuffer(1024)
	if cap(buf2) < 1024 {
		t.Errorf("Buffer capacity too small on reuse: %d", cap(buf2))
	}
}

func TestGetBool(t *testing.T) {
	msg := GetMessage()
	msg.AddField(TagPossDupFlag, "Y")
	msg.AddField(TagResetSeqNumFlag, "N")
	
	val, ok := msg.GetBool(TagPossDupFlag)
	if !ok || val != true {
		t.Errorf("Expected true, got %v, ok=%v", val, ok)
	}
	
	val, ok = msg.GetBool(TagResetSeqNumFlag)
	if !ok || val != false {
		t.Errorf("Expected false, got %v, ok=%v", val, ok)
	}
	
	val, ok = msg.GetBool(TagSide)
	if ok {
		t.Error("Expected false for missing field")
	}
	
	PutMessage(msg)
}

func TestErrorHandling(t *testing.T) {
	_, err := Parse([]byte("invalid"))
	if err == nil {
		t.Error("Expected error for invalid format")
	}
	
	_, err = Parse([]byte(""))
	if err == nil {
		t.Error("Expected error for empty input")
	}
}

func TestGetUint(t *testing.T) {
	msg := GetMessage()
	msg.AddField(TagMsgSeqNum, "12345")
	msg.AddField(TagQuantity, "100")
	
	val, ok := msg.GetUint(TagMsgSeqNum)
	if !ok || val != 12345 {
		t.Errorf("Expected 12345, got %d, ok=%v", val, ok)
	}
	
	PutMessage(msg)
}
