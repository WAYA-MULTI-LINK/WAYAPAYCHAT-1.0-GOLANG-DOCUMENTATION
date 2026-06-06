package wayapay_test

import wayapay "github.com/wayapaychat/wayapay-go/src/wayapay"

func validPayout() wayapay.PayoutRequest {
	return wayapay.PayoutRequest{
		Amount:        5000,
		Currency:      "NGN",
		AccountNumber: "0123456789",
		BankCode:      "044",
		AccountName:   "JOHN DOE",
		Reference:     "REF-001",
		Narration:     "Test payout",
	}
}

func validVerify() wayapay.VerifyAccountRequest {
	return wayapay.VerifyAccountRequest{
		AccountNumber: "0123456789",
		BankCode:      "044",
		EnquiryType:   wayapay.EnquiryTypeOthers,
	}
}

func validCollect() wayapay.CollectRequest {
	return wayapay.CollectRequest{
		PaymentLinkType: wayapay.PaymentLinkTypeOneTime,
		PaymentLinkName: "Order #1234",
		Description:     "Order #1234 - 2 items",
		PayableAmount:   1500,
		Currency:        "NGN",
		RedirectLink:    "https://merchant.example.com/callback",
	}
}
