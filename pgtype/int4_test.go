package pgtype_test

import (
	"bytes"
	"math"
	"testing"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgio"
	"github.com/jackc/pgx/pgtype"
)

func TestInt4Transcode(t *testing.T) {
	conn := mustConnectPgx(t)
	defer mustClose(t, conn)

	ps, err := conn.Prepare("test", "select $1::int4")
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		result pgtype.Int4
	}{
		{result: pgtype.Int4(math.MinInt32)},
		{result: pgtype.Int4(-1)},
		{result: pgtype.Int4(0)},
		{result: pgtype.Int4(1)},
		{result: pgtype.Int4(math.MaxInt32)},
	}

	ps.FieldDescriptions[0].FormatCode = pgx.TextFormatCode
	for i, tt := range tests {
		inputBuf := &bytes.Buffer{}
		err = tt.result.EncodeText(inputBuf)
		if err != nil {
			t.Errorf("TextFormat %d: %v", i, err)
		}

		var s string
		err := conn.QueryRow("test", string(inputBuf.Bytes()[4:])).Scan(&s)
		if err != nil {
			t.Errorf("TextFormat %d: %v", i, err)
		}

		outputBuf := &bytes.Buffer{}
		pgio.WriteInt32(outputBuf, int32(len(s)))
		outputBuf.WriteString(s)
		var r pgtype.Int4
		err = r.DecodeText(outputBuf)
		if err != nil {
			t.Errorf("TextFormat %d: %v", i, err)
		}

		if r != tt.result {
			t.Errorf("TextFormat %d: expected %v, got %v", i, tt.result, r)
		}
	}

	ps.FieldDescriptions[0].FormatCode = pgx.BinaryFormatCode
	for i, tt := range tests {
		inputBuf := &bytes.Buffer{}
		err = tt.result.EncodeBinary(inputBuf)
		if err != nil {
			t.Errorf("BinaryFormat %d: %v", i, err)
		}

		var buf []byte
		err := conn.QueryRow("test", inputBuf.Bytes()[4:]).Scan(&buf)
		if err != nil {
			t.Errorf("BinaryFormat %d: %v", i, err)
		}

		outputBuf := &bytes.Buffer{}
		pgio.WriteInt32(outputBuf, int32(len(buf)))
		outputBuf.Write(buf)
		var r pgtype.Int4
		err = r.DecodeBinary(outputBuf)
		if err != nil {
			t.Errorf("BinaryFormat %d: %v", i, err)
		}

		if r != tt.result {
			t.Errorf("BinaryFormat %d: expected %v, got %v", i, tt.result, r)
		}
	}
}