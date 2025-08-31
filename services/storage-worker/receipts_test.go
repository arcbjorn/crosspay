package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestHandleGenerateReceipt(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	// Initialize storage service for tests
	initializeStorageService()
	
	router := gin.New()
	router.POST("/receipts", handleGenerateReceipt)

	t.Run("should generate receipt successfully", func(t *testing.T) {
		req := GenerateReceiptRequest{
			PaymentID: 123,
			Format:    "json",
			Language:  "en",
		}

		reqBody, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		httpReq, _ := http.NewRequest("POST", "/receipts", bytes.NewBuffer(reqBody))
		httpReq.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusOK, w.Code)

		var response GenerateReceiptResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.NotEmpty(t, response.ReceiptID)
		assert.NotEmpty(t, response.CID)
		assert.Equal(t, "json", response.Format)
		assert.Greater(t, response.Size, int64(0))
	})

	t.Run("should generate PDF receipt", func(t *testing.T) {
		req := GenerateReceiptRequest{
			PaymentID: 456,
			Format:    "pdf",
			Language:  "en",
		}

		reqBody, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		httpReq, _ := http.NewRequest("POST", "/receipts", bytes.NewBuffer(reqBody))
		httpReq.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusOK, w.Code)

		var response GenerateReceiptResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "pdf", response.Format)
	})

	t.Run("should handle invalid request format", func(t *testing.T) {
		w := httptest.NewRecorder()
		httpReq, _ := http.NewRequest("POST", "/receipts", bytes.NewBuffer([]byte("invalid json")))
		httpReq.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestHandleDownloadReceipt(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/receipts/:id", handleDownloadReceipt)

	t.Run("should download receipt successfully", func(t *testing.T) {
		w := httptest.NewRecorder()
		httpReq, _ := http.NewRequest("GET", "/receipts/rcpt_123_1640995200", nil)

		router.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.NotEmpty(t, w.Body.String())
	})

	t.Run("should handle missing receipt ID", func(t *testing.T) {
		w := httptest.NewRecorder()
		httpReq, _ := http.NewRequest("GET", "/receipts/", nil)

		router.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestHandleVerifyReceipt(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/verify/:cid", handleVerifyReceipt)

	t.Run("should verify receipt successfully", func(t *testing.T) {
		w := httptest.NewRecorder()
		httpReq, _ := http.NewRequest("GET", "/verify/bafybeigtest123", nil)

		router.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Contains(t, response, "cid")
		assert.Contains(t, response, "valid")
		assert.Contains(t, response, "payment_id")
	})

	t.Run("should handle missing CID", func(t *testing.T) {
		w := httptest.NewRecorder()
		httpReq, _ := http.NewRequest("GET", "/verify/", nil)

		router.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestFetchPaymentData(t *testing.T) {
	t.Run("should fetch payment data", func(t *testing.T) {
		payment, err := fetchPaymentData(123)

		assert.NoError(t, err)
		assert.NotNil(t, payment)
		assert.Equal(t, uint64(123), payment.ID)
		assert.Equal(t, "completed", payment.Status)
		assert.Equal(t, 1135, payment.ChainID)
		assert.NotEmpty(t, payment.Sender)
		assert.NotEmpty(t, payment.Recipient)
	})
}

func TestGenerateReceipt(t *testing.T) {
	mockPayment := &PaymentData{
		ID:           123,
		Sender:       "0x1234567890123456789012345678901234567890",
		SenderENS:    "alice.eth",
		Recipient:    "0x0987654321098765432109876543210987654321",
		RecipientENS: "bob.eth",
		Token:        "0x0000000000000000000000000000000000000000",
		Amount:       "1000000000000000000",
		Fee:          "1000000000000000",
		Status:       "completed",
		ChainID:      1135,
		OraclePrice:  "2500.00",
	}

	t.Run("should generate JSON receipt", func(t *testing.T) {
		receipt, err := generateReceipt(mockPayment, "json", "en")

		assert.NoError(t, err)
		assert.NotNil(t, receipt)
		assert.Equal(t, mockPayment.ID, receipt.Payment.ID)
		assert.Equal(t, "1.0", receipt.Version)
		assert.Equal(t, "json", receipt.Format)
		assert.NotEmpty(t, receipt.Signature)
		assert.Contains(t, receipt.Metadata, "language")
		assert.Equal(t, "en", receipt.Metadata["language"])
	})

	t.Run("should generate PDF receipt", func(t *testing.T) {
		receipt, err := generateReceipt(mockPayment, "pdf", "en")

		assert.NoError(t, err)
		assert.NotNil(t, receipt)
		assert.Equal(t, "pdf", receipt.Format)
	})
}

func TestGeneratePDFReceipt(t *testing.T) {
	receipt := &Receipt{
		Payment: PaymentData{
			ID:           123,
			SenderENS:    "alice.eth",
			RecipientENS: "bob.eth",
			Amount:       "1000000000000000000",
			Status:       "completed",
			ChainID:      1135,
		},
		GeneratedAt: time.Now(),
		Signature:   "sig_test123",
	}

	t.Run("should generate PDF content", func(t *testing.T) {
		pdfData, err := generatePDFReceipt(receipt)

		assert.NoError(t, err)
		assert.NotNil(t, pdfData)
		assert.Contains(t, string(pdfData), "CrossPay Payment Receipt")
		assert.Contains(t, string(pdfData), "alice.eth")
		assert.Contains(t, string(pdfData), "bob.eth")
		assert.Contains(t, string(pdfData), "completed")
	})
}

func TestSignAndVerifyReceipt(t *testing.T) {
	receipt := &Receipt{
		Payment: PaymentData{
			ID:     123,
			Amount: "1000000000000000000",
		},
		GeneratedAt: time.Now(),
	}

	t.Run("should sign receipt", func(t *testing.T) {
		signature, err := signReceipt(receipt)

		assert.NoError(t, err)
		assert.NotEmpty(t, signature)
		assert.Contains(t, signature, "sig_")
	})

	t.Run("should verify receipt signature", func(t *testing.T) {
		signature, err := signReceipt(receipt)
		assert.NoError(t, err)

		receipt.Signature = signature
		isValid := verifyReceiptSignature(*receipt)

		assert.True(t, isValid)
	})

	t.Run("should fail verification with invalid signature", func(t *testing.T) {
		receipt.Signature = "invalid_signature"
		isValid := verifyReceiptSignature(*receipt)

		assert.False(t, isValid)
	})
}

func TestGetCIDFromReceiptID(t *testing.T) {
	t.Run("should extract CID from receipt ID", func(t *testing.T) {
		cid, err := getCIDFromReceiptID("rcpt_123_1640995200")

		assert.NoError(t, err)
		assert.NotEmpty(t, cid)
		assert.Contains(t, cid, "bafybeigmock")
	})
}

func TestGetNetworkName(t *testing.T) {
	testCases := []struct {
		chainID      int
		expectedName string
	}{
		{1135, "Lisk"},
		{84532, "Base Sepolia"},
		{5115, "Citrea Testnet"},
		{999999, "Unknown"},
	}

	for _, tc := range testCases {
		t.Run(tc.expectedName, func(t *testing.T) {
			name := getNetworkName(tc.chainID)
			assert.Equal(t, tc.expectedName, name)
		})
	}
}