package wayaquick_test

import wayaquick "github.com/WAYA-MULTI-LINK/WAYAPAYCHAT-1.0-GOLANG-DOCUMENTATION/src/wayaquick"

func validPayout() wayaquick.PayoutRequest {
	return wayaquick.PayoutRequest{
		Amount:        5000,
		Currency:      "NGN",
		AccountNumber: "0123456789",
		BankCode:      "044",
		AccountName:   "JOHN DOE",
		Reference:     "REF-001",
		Narration:     "Test payout",
	}
}

func validVerify() wayaquick.VerifyAccountRequest {
	return wayaquick.VerifyAccountRequest{
		AccountNumber: "0123456789",
		BankCode:      "044",
		EnquiryType:   wayaquick.EnquiryTypeOthers,
	}
}

func validCollect() wayaquick.CollectRequest {
	return wayaquick.CollectRequest{
		PaymentLinkType: wayaquick.PaymentLinkTypeOneTime,
		PaymentLinkName: "Order #1234",
		Description:     "Order #1234 - 2 items",
		PayableAmount:   1500,
		Currency:        "NGN",
		RedirectLink:    "https://merchant.example.com/callback",
	}
}
