package tests

import (
	"mailfinger/query/pop3"
	"mailfinger/query/pop3s"
	"mailfinger/query/smtp"
	"testing"
)

func TestSingle(t *testing.T) {

	testCases := []string{
		"exmail.qq.com",
	}

	for _, m := range testCases {
		smtp.DoQuery(m)
		pop3.DoQuery(m)
		pop3s.DoQuery(m)
	}
}
