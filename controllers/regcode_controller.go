package controllers

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"strings"
	"text/template"

	svc "cloudminds.com/harix/cc-server/services"
	"github.com/gin-gonic/gin"
	"github.com/sfreiberg/gotwilio"
	"go.mongodb.org/mongo-driver/mongo"
)

// RegCodeEmailContent - as is
type RegCodeEmailContent struct {
	Name     string
	PhoneNum string
	RegCode  string
}

// RegCodeSMSContent - as is
type RegCodeSMSContent struct {
	Name    string
	RegCode string
}

// SendRegCodeForm - as is
type SendRegCodeForm struct {
	ID          string `json:"id"`
	FirstName   string `json:"first_name"`
	PhoneNum    string `json:"phone_num"`
	ToEmailAddr string `json:"to_email_addr"`
}

// GetManyRegCodes - For Debug Purpose
func (s *CCServer) GetManyRegCodes(c *gin.Context) {
	cursor, err := svc.GetManyRegCodes()

	if err != nil {
		log.Printf("Error while getting all regCodes - %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  http.StatusInternalServerError,
			"message": "Something went wrong",
		})
		return
	}
	// Iterate through the returned cursor
	regCodes := []svc.RegCode{}
	if err = cursor.All(context.TODO(), &regCodes); err != nil {
		panic(err)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "All Reg Codes",
		"data":    regCodes,
	})
	return
}

// GetRegCodeByMemberID - as is
func (s *CCServer) GetRegCodeByMemberID(c *gin.Context) {
	var queryParams svc.GetRegCodeParams
	queryParams.MemberID = c.DefaultQuery("memberID", "000000000000000000000000")
	// Iterate through the returned cursor
	regCode := svc.RegCode{}
	err := svc.GetRegCodeByMemberID(queryParams.MemberID).Decode(&regCode)

	if err != nil {
		log.Printf("Error while getting regcode by GuardianID - %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  http.StatusInternalServerError,
			"message": "Something went wrong",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Found Reg Code by Guardian ID",
		"data":    regCode,
	})
	return
}

// SendRegCodeWithEmail - as is
func (s *CCServer) SendRegCodeWithEmail(c *gin.Context) {
	sendRegCodeForm := SendRegCodeForm{}
	c.BindJSON(&sendRegCodeForm)
	// log.Printf("SendRegCodeWithEmail, receive POST body - %v\n", sendRegCodeForm)

	// Get RegCode
	regCode := svc.RegCode{}
	err := svc.GetRegCodeByID(sendRegCodeForm.ID).Decode(&regCode)

	if err != nil {
		log.Printf("Error while getting regcode by ID - %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  http.StatusInternalServerError,
			"message": "Something went wrong",
		})
		return
	}

	// Send Email
	emailContent := RegCodeEmailContent{
		Name:     sendRegCodeForm.FirstName,
		PhoneNum: sendRegCodeForm.PhoneNum,
		RegCode:  regCode.RegCode,
	}
	err = s.handleSendRegCodeWithEmail(emailContent, sendRegCodeForm.ToEmailAddr)
	if err != nil {
		log.Printf("SendRegCodeWithEmail - Error while Sending Email - %v\n", err)
	}
	log.Println("Email Sent!")
	c.JSON(http.StatusOK, gin.H{
		"message": "Email Sent!",
	})

	s.sendRegCodePostProcessing(c, sendRegCodeForm.PhoneNum)
}

// SendRegCodeWithSMS - as is
func (s *CCServer) SendRegCodeWithSMS(c *gin.Context) {
	sendRegCodeForm := SendRegCodeForm{}
	c.BindJSON(&sendRegCodeForm)

	// Get RegCode
	regCode := svc.RegCode{}
	err := svc.GetRegCodeByID(sendRegCodeForm.ID).Decode(&regCode)

	if err != nil {
		log.Printf("Error while getting regcode by ID - %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  http.StatusInternalServerError,
			"message": "Something went wrong",
		})
		return
	}

	// Send SMS
	smsContent := RegCodeSMSContent{
		Name:    sendRegCodeForm.FirstName,
		RegCode: regCode.RegCode,
	}

	s.handleSendRegCodeWithSMS(smsContent, sendRegCodeForm.PhoneNum)

	log.Println("SMS Sent!")
	c.JSON(http.StatusOK, gin.H{
		"message": "SMS Sent!",
	})

	s.sendRegCodePostProcessing(c, sendRegCodeForm.PhoneNum)
}

// SendRegCodeWithEmail - as is
func (s *CCServer) handleSendRegCodeWithEmail(ec RegCodeEmailContent, toEmailAddr string) error {

	log.Printf("SendRegCodeWithEmail - Email Config %v\n", s.Config.EmailConf)
	emailConfig := s.Config.EmailConf
	auth := smtp.PlainAuth(
		"",
		emailConfig.FromEmailAddr,
		emailConfig.FromEmailPswd,
		emailConfig.SMTPHost,
	)
	t, _ := template.ParseFiles("templates/reg_code_email_template.html")

	var body bytes.Buffer

	mimeHeaders := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	body.Write([]byte(fmt.Sprintf("Subject: This is a test subject \n%s\n\n", mimeHeaders)))

	t.Execute(&body, ec)
	// Sending email.
	err := smtp.SendMail(
		emailConfig.SMTPHost+":"+emailConfig.SMTPPort,
		auth,
		emailConfig.FromEmailAddr,
		[]string{toEmailAddr},
		body.Bytes(),
	)
	return err
}

// SendRegCodeWithSMS - as is
func (s *CCServer) handleSendRegCodeWithSMS(sc RegCodeSMSContent, toPhoneNum string) {
	log.Printf("SendRegCodeWithSMS - SMS Config %v\n", s.Config.SMSConf)
	smsConfig := s.Config.SMSConf
	twilio := gotwilio.NewTwilioClient(smsConfig.AccountSid, smsConfig.AuthToken)

	// convert db phonenum to twilio from ("xxx-xxx-xxxx" to "+1xxxxxxxxxx")
	toPhoneNum = "+1" + strings.ReplaceAll(toPhoneNum, "-", "")
	message := "Hi " + sc.Name + "! Thank you for using the Check-in/Check-out screening system! " +
		"Your Registration Code is: " + sc.RegCode + ". " +
		"The Mobile WebApp can be accessed at: " + s.Config.ServerAddr + "/mobile/login"
	twilio.SendSMS(smsConfig.FromPhoneNum, toPhoneNum, message, "", "")
}

func (s *CCServer) sendRegCodePostProcessing(c *gin.Context, phoneNum string) {
	// Update Member Status
	// Get Member
	err := svc.SetMemberRegCodeSentByPhoneNum(phoneNum)
	if err != nil {
		// When no member found, return failed
		if err == mongo.ErrNoDocuments {

			log.Printf("Member Not Found, Status Changing Failed - %v\n", err)
			return
		}
		log.Printf("Error while getting Member given PhoneNum - %v\n", err)
		return
	}

}
