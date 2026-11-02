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
		TransactionID: "TXN-V2-20260604-01",
		Amount:        "1500.00",
		Currency:      "NGN",
		FirstName:     "John",
		LastName:      "Doe",
		Email:         "john@example.com",
		Phone:         "08012345678",
		Description:   "Order #1234",
	}
}
