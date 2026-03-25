package compare

import (
	"bytes"
	"os"
	"testing"
	"time"

	fixzero "github.com/fixzero/fixzero"
	"github.com/quickfixgo/quickfix"
)

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
		if err != nil {
			b.Fatal(err)
		}
		result, err = msg.Body.GetString(11)
		if err != nil {
			b.Fatal(err)
		}
		result, err = msg.Body.GetString(54)
		if err != nil {
			b.Fatal(err)
		}
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

func BenchmarkSessionStateCreation(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		ss := fixzero.GetSessionState()
		ss.SessionID = fixzero.SessionID{
			BeginString:  "FIX.4.4",
			SenderCompID: "SENDER",
			TargetCompID: "TARGET",
		}
		ss.HeartBtInt = 30
		ss.OutMsgSeqNum = 1
		ss.InMsgSeqNum = 1
		fixzero.PutSessionState(ss)
	}
}

func BenchmarkSessionSeqNumIncr(b *testing.B) {
	ss := fixzero.GetSessionState()
	ss.HeartBtInt = 30

	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		ss.IncrOutMsgSeqNum()
		ss.IncrInMsgSeqNum()
	}
}

func BenchmarkSessionHeartbeatCheck(b *testing.B) {
	ss := fixzero.GetSessionState()
	ss.IsLogon = true
	ss.HeartBtInt = 30
	ss.LastSentTime = time.Now().Add(-25 * time.Second)
	ss.LastRecvTime = time.Now().Add(-25 * time.Second)

	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = ss.NeedHeartbeat()
		_ = ss.CheckHeartbeat()
	}
}

func BenchmarkFixZeroValidateField(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = fixzero.ValidateType("AAPL", "STRING")
	}
}

func BenchmarkFixZeroValidateFieldInt(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = fixzero.ValidateType("100", "INT")
	}
}

func BenchmarkFixZeroValidateFieldFloat(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = fixzero.ValidateType("150.50", "FLOAT")
	}
}

func BenchmarkDataDictionaryLoad(b *testing.B) {
	dd := fixzero.NewDataDictionary()

	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = dd.ParseXML([]byte(testDDXML))
	}
}

func BenchmarkNewOrderSingleEncode(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		msg := fixzero.NewBuilder().
			BeginString("FIX.4.4").
			MsgType("D").
			Set(fixzero.TagMsgSeqNum, "1").
			Set(fixzero.TagSenderCompID, "SENDER").
			Set(fixzero.TagTargetCompID, "TARGET").
			Set(fixzero.TagClOrdID, "ORDER123").
			Set(fixzero.TagSymbol, "AAPL").
			Set(fixzero.TagSide, "1").
			Set(fixzero.TagOrdType, "2").
			SetInt(fixzero.TagQuantity, 100).
			SetFloat(fixzero.TagPrice, 150.50).
			Set(fixzero.TagTimeInForce, "0").
			Build()

		fixzero.PutMessage(msg)
	}
}

func BenchmarkNewOrderSingleDecode(b *testing.B) {
	nosData := []byte("8=FIX.4.4|9=112|35=D|49=SENDER|56=TARGET|11=ORDER123|55=AAPL|54=1|40=2|38=100|44=150.50|59=0|10=000|")
	nosData = bytes.ReplaceAll(nosData, []byte("|"), []byte{1})

	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		msg, err := fixzero.Parse(nosData)
		if err != nil {
			b.Fatal(err)
		}
		_ = msg.GetString(fixzero.TagClOrdID)
		_ = msg.GetString(fixzero.TagSymbol)
		_ = msg.GetString(fixzero.TagSide)
		fixzero.PutMessage(msg)
	}
}

func BenchmarkExecutionReportEncode(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		msg := fixzero.NewBuilder().
			BeginString("FIX.4.4").
			MsgType("8").
			Set(fixzero.TagMsgSeqNum, "1").
			Set(fixzero.TagSenderCompID, "SENDER").
			Set(fixzero.TagTargetCompID, "TARGET").
			Set(fixzero.TagExecID, "EXEC123").
			Set(fixzero.TagExecRefID, "ORDER123").
			Set(fixzero.TagExecType, "0").
			Set(fixzero.TagOrdStatus, "0").
			Set(fixzero.TagSymbol, "AAPL").
			Set(fixzero.TagSide, "1").
			SetInt(fixzero.TagQuantity, 100).
			SetFloat(fixzero.TagLastPx, 150.50).
			SetFloat(fixzero.TagAvgPx, 150.50).
			Set(fixzero.TagTimeInForce, "0").
			Build()

		fixzero.PutMessage(msg)
	}
}

func BenchmarkExecutionReportDecode(b *testing.B) {
	erData := []byte("8=FIX.4.4|9=150|35=8|49=SENDER|56=TARGET|37=ORDER123|17=EXEC123|150=0|39=0|55=AAPL|54=1|38=100|44=150.50|6=150.50|10=000|")
	erData = bytes.ReplaceAll(erData, []byte("|"), []byte{1})

	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		msg, err := fixzero.Parse(erData)
		if err != nil {
			b.Fatal(err)
		}
		_ = msg.GetString(fixzero.TagExecID)
		_ = msg.GetString(fixzero.TagOrdStatus)
		fixzero.PutMessage(msg)
	}
}

func BenchmarkMemoryStoreSave(b *testing.B) {
	store := fixzero.NewMemoryStore()
	msgData := []byte("8=FIX.4.4|35=D|11=ORDER123|55=AAPL|")

	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = store.SaveMessage(i+1, msgData)
	}
}

func BenchmarkMemoryStoreGet(b *testing.B) {
	store := fixzero.NewMemoryStore()
	msgData := []byte("8=FIX.4.4|35=D|11=ORDER123|55=AAPL|")
	for i := 1; i <= 100; i++ {
		store.SaveMessage(i, msgData)
	}

	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = store.GetMessage((i % 100) + 1)
	}
}

func BenchmarkFileStoreSave(b *testing.B) {
	tmpDir, err := os.MkdirTemp("", "bench_filestore")
	if err != nil {
		b.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	store, err := fixzero.NewFileStore(tmpDir)
	if err != nil {
		b.Fatal(err)
	}

	msgData := []byte("8=FIX.4.4|35=D|11=ORDER123|55=AAPL|")

	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = store.SaveMessage(i+1, msgData)
	}

	store.Close()
}

func BenchmarkFileStoreGet(b *testing.B) {
	tmpDir, err := os.MkdirTemp("", "bench_filestore")
	if err != nil {
		b.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	store, err := fixzero.NewFileStore(tmpDir)
	if err != nil {
		b.Fatal(err)
	}

	msgData := []byte("8=FIX.4.4|35=D|11=ORDER123|55=AAPL|")
	for i := 1; i <= 100; i++ {
		store.SaveMessage(i, msgData)
	}

	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = store.GetMessage((i % 100) + 1)
	}

	store.Close()
}

func BenchmarkNullLog(b *testing.B) {
	log := &fixzero.NullLog{}
	testMsg := "8=FIX.4.4|35=D|11=ORDER123|55=AAPL|"

	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		log.OnIncoming(testMsg)
		log.OnOutgoing(testMsg)
		log.OnEvent("test event")
		log.OnError("test error")
	}
}

func BenchmarkScreenLog(b *testing.B) {
	log := fixzero.NewScreenLog(false)
	testMsg := "8=FIX.4.4|35=D|11=ORDER123|55=AAPL|"

	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		log.OnIncoming(testMsg)
		log.OnOutgoing(testMsg)
		log.OnEvent("test event")
		log.OnError("test error")
	}
}

func BenchmarkFileLog(b *testing.B) {
	tmpDir, err := os.MkdirTemp("", "bench_filelog")
	if err != nil {
		b.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	log, err := fixzero.NewFileLog(fixzero.WithLogDir(tmpDir))
	if err != nil {
		b.Fatal(err)
	}
	defer log.Close()

	testMsg := "8=FIX.4.4|35=D|11=ORDER123|55=AAPL|"

	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		log.OnIncoming(testMsg)
		log.OnOutgoing(testMsg)
		log.OnEvent("test event")
		log.OnError("test error")
	}
}

func BenchmarkCountGroups(b *testing.B) {
	nosData := []byte("8=FIX.4.4|9=200|35=D|11=ORDER123|55=AAPL|54=1|40=2|38=100|44=150.50|555=2|600=GRPTAG1|601=GRPTAG2|600=GRPTAG3|601=GRPTAG4|10=000|")
	msg, _ := fixzero.Parse(nosData)

	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = fixzero.CountGroupsByTag(msg, 555)
	}
}

func BenchmarkIterateGroups(b *testing.B) {
	nosData := []byte("8=FIX.4.4|9=200|35=D|11=ORDER123|55=AAPL|54=1|40=2|38=100|44=150.50|555=2|600=GRPTAG1|601=GRPTAG2|600=GRPTAG3|601=GRPTAG4|10=000|")
	msg, _ := fixzero.Parse(nosData)

	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		fixzero.IterateGroups(msg, 555, func(groupIndex int, fields []fixzero.Field) bool {
			_ = groupIndex
			_ = len(fields)
			return true
		})
	}
}

func BenchmarkIterateGroupsGetGroup(b *testing.B) {
	nosData := []byte("8=FIX.4.4|9=200|35=D|11=ORDER123|55=AAPL|54=1|40=2|38=100|44=150.50|555=2|600=GRPTAG1|601=GRPTAG2|600=GRPTAG3|601=GRPTAG4|10=000|")
	msg, _ := fixzero.Parse(nosData)

	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		for j := 0; j < 2; j++ {
			_ = fixzero.GetGroupByTag(msg, 555, j)
		}
	}
}

func BenchmarkInitiatorCreate(b *testing.B) {
	app := &fixzero.ApplicationFuncs{}
	settings := &fixzero.Settings{
		GlobalSessions: map[fixzero.SessionID]*fixzero.SessionSettings{
			{
				BeginString:  "FIX.4.4",
				SenderCompID: "SENDER",
				TargetCompID: "TARGET",
			}: {
				Host:              "localhost",
				Port:              9876,
				HeartbeatInterval: 30,
			},
		},
	}

	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = fixzero.NewInitiator(app, settings)
	}
}

func BenchmarkAcceptorCreate(b *testing.B) {
	app := &fixzero.ApplicationFuncs{}
	settings := &fixzero.Settings{
		GlobalSessions: map[fixzero.SessionID]*fixzero.SessionSettings{
			{
				BeginString:  "FIX.4.4",
				SenderCompID: "SENDER",
				TargetCompID: "TARGET",
			}: {
				Host: "localhost",
				Port: 9876,
			},
		},
	}

	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = fixzero.NewAcceptor(app, settings)
	}
}

func BenchmarkMessagePoolGet(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		msg := fixzero.GetMessage()
		fixzero.PutMessage(msg)
	}
}

func BenchmarkQuickFIXMessageNew(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		msg := quickfix.NewMessage()
		_ = msg
	}
}

func BenchmarkQuickFIXSetFields(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		msg := quickfix.NewMessage()
		msg.Header.SetString(8, "FIX.4.4")
		msg.Header.SetString(35, "D")
		msg.Header.SetString(49, "SENDER")
		msg.Header.SetString(56, "TARGET")
		msg.Header.SetInt(34, 1)
		msg.Body.SetString(11, "ORDER123")
		msg.Body.SetString(55, "AAPL")
		msg.Body.SetString(54, "1")
		msg.Body.SetString(40, "2")
		msg.Body.SetInt(38, 100)
		msg.Body.SetString(44, "150.50")
		msg.Body.SetString(59, "0")
	}
}

func BenchmarkQuickFIXGetFields(b *testing.B) {
	msg := quickfix.NewMessage()
	msg.Header.SetString(8, "FIX.4.4")
	msg.Header.SetString(35, "D")
	msg.Header.SetString(49, "SENDER")
	msg.Header.SetString(56, "TARGET")
	msg.Body.SetString(11, "ORDER123")
	msg.Body.SetString(55, "AAPL")
	msg.Body.SetString(54, "1")
	msg.Body.SetString(40, "2")
	msg.Body.SetInt(38, 100)
	msg.Body.SetString(44, "150.50")

	b.ResetTimer()
	b.ReportAllocs()

	var result string
	var err error

	for i := 0; i < b.N; i++ {
		result, _ = msg.Header.GetString(49)
		result, _ = msg.Header.GetString(56)
		result, _ = msg.Body.GetString(11)
		result, _ = msg.Body.GetString(55)
		result, _ = msg.Body.GetString(54)
		result, _ = msg.Body.GetString(40)
	}
	_ = result
	_ = err
}

func BenchmarkQuickFIXSerializeLarge(b *testing.B) {
	msg := quickfix.NewMessage()
	msg.Header.SetString(8, "FIX.4.4")
	msg.Header.SetString(35, "D")
	msg.Header.SetString(49, "SENDER")
	msg.Header.SetString(56, "TARGET")
	msg.Header.SetInt(34, 1)
	msg.Body.SetString(11, "ORDER123")
	msg.Body.SetString(55, "AAPL")
	msg.Body.SetString(54, "1")
	msg.Body.SetString(40, "2")
	msg.Body.SetInt(38, 100)
	msg.Body.SetString(44, "150.50")
	msg.Body.SetString(59, "0")
	msg.Body.SetString(21, "1")
	msg.Body.SetString(203, "0")
	msg.Body.SetString(205, "4")
	msg.Body.SetString(207, "NYSE")
	msg.Body.SetString(216, "1")
	msg.Body.SetString(110, "0")
	msg.Body.SetString(111, "100")

	b.ReportAllocs()
	b.ResetTimer()

	var result []byte

	for i := 0; i < b.N; i++ {
		result = msg.Bytes()
	}
	_ = result
}

func BenchmarkFixZeroSerializeLarge(b *testing.B) {
	msg := fixzero.NewBuilder().
		BeginString("FIX.4.4").
		MsgType("D").
		Set(fixzero.TagSenderCompID, "SENDER").
		Set(fixzero.TagTargetCompID, "TARGET").
		Set(fixzero.TagMsgSeqNum, "1").
		Set(fixzero.TagClOrdID, "ORDER123").
		Set(fixzero.TagSymbol, "AAPL").
		Set(fixzero.TagSide, "1").
		Set(fixzero.TagOrdType, "2").
		SetInt(fixzero.TagQuantity, 100).
		SetFloat(fixzero.TagPrice, 150.50).
		Set(fixzero.TagTimeInForce, "0").
		Set(fixzero.TagHandlInst, "1").
		Set(fixzero.TagExecInst, "0").
		Build()

	b.ReportAllocs()
	b.ResetTimer()

	var result []byte
	for i := 0; i < b.N; i++ {
		result = fixzero.Serialize(msg)
	}
	_ = result
}

func BenchmarkSessionStatePool(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		ss := fixzero.GetSessionState()
		ss.SessionID = fixzero.SessionID{
			BeginString:  "FIX.4.4",
			SenderCompID: "SENDER",
			TargetCompID: "TARGET",
		}
		ss.HeartBtInt = 30
		fixzero.PutSessionState(ss)
	}
}

func BenchmarkSerializeState(b *testing.B) {
	ss := fixzero.GetSessionState()
	ss.SessionID = fixzero.SessionID{
		BeginString:  "FIX.4.4",
		SenderCompID: "SENDER",
		TargetCompID: "TARGET",
	}
	ss.HeartBtInt = 30
	ss.OutMsgSeqNum = 1
	ss.InMsgSeqNum = 1
	ss.IsLogon = true

	b.ReportAllocs()

	var result string
	for i := 0; i < b.N; i++ {
		result = ss.SessionID.String()
	}
	_ = result
}

func BenchmarkRepeatingGroupValidation(b *testing.B) {
	nosData := []byte("8=FIX.4.4|9=250|35=D|11=ORDER123|55=AAPL|54=1|40=2|38=100|44=150.50|555=3|600=GRPTAG1|601=GRPTAG2|600=GRPTAG3|601=GRPTAG4|600=GRPTAG5|601=GRPTAG6|10=000|")
	msg, _ := fixzero.Parse(nosData)

	groupDef := fixzero.NewGroupDef(555, 555, []int{600, 601})

	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = fixzero.ValidateGroup(msg, groupDef)
	}
}

func BenchmarkMemoryStoreSeqNumOps(b *testing.B) {
	store := fixzero.NewMemoryStore()

	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		store.SetNextSenderSeqNum(i + 1)
		store.SetNextTargetSeqNum(i + 1)
		store.IncrNextSenderSeqNum()
		store.IncrNextTargetSeqNum()
		_ = store.GetNextSenderSeqNum()
		_ = store.GetNextTargetSeqNum()
	}
}

var testDDXML = `<?xml version="1.0" encoding="UTF-8"?>
<fix version="4.4">
  <header>
    <field name="BeginString" required="Y"/>
    <field name="BodyLength" required="Y"/>
    <field name="MsgType" required="Y"/>
    <field name="SenderCompID" required="Y"/>
    <field name="TargetCompID" required="Y"/>
    <field name="MsgSeqNum" required="Y"/>
  </header>
  <messages>
    <msg name="NewOrderSingle" msgtype="D">
      <field name="ClOrdID" required="Y"/>
      <field name="Symbol" required="Y"/>
      <field name="Side" required="Y"/>
      <field name="TransactTime" required="Y"/>
      <field name="OrdType" required="Y"/>
      <field name="OrderQty" required="N"/>
      <field name="Price" required="N"/>
    </msg>
    <msg name="ExecutionReport" msgtype="8">
      <field name="OrderID" required="Y"/>
      <field name="ExecID" required="Y"/>
      <field name="ExecType" required="Y"/>
      <field name="OrdStatus" required="Y"/>
      <field name="Symbol" required="Y"/>
      <field name="Side" required="Y"/>
      <field name="OrderQty" required="N"/>
    </msg>
  </messages>
</fix>`
