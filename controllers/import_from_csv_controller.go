package controllers

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"net/http"

	svc "cloudminds.com/harix/cc-server/services"
	"github.com/gin-gonic/gin"
)

// ImportManyTags - batch import tags through CSV file
func (s *CCServer) ImportManyTags(c *gin.Context) {

	instID := c.DefaultQuery("instID", "000000000000000000000000")

	file, header, err := c.Request.FormFile("file")
	defer file.Close()
	if err != nil {
		log.Panic(err)
	}

	// log.Println(file.Filename)

	// err = c.SaveUploadedFile(file, "saved/"+file.Filename)
	// out, err := os.Create()
	// TODO - Update/Merge the Existing Tag with Form Data
	r := csv.NewReader(file)
	count := 0
	for {
		record, err := r.Read()
		if count == 0 {
			count++
			continue
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		if len(record) < 6 {
			log.Panic("record fields are too few!")
		}

		// Check if Tag already exists
		countTagParams := svc.CountTagParams{
			TagString: record[0],
			InstID:    instID,
		}
		tagCount, _ := svc.CountTag(countTagParams)
		if tagCount > 0 {
			log.Printf("Tag %v already exists!\n", record[0])
			continue
		}

		tRegForm := svc.TagRegForm{
			InstID:    instID,
			TagString: record[0],
			FirstName: record[1],
			LastName:  record[2],
			Group:     record[3],
			PhoneNum:  record[4],
			Email:     record[5],
		}
		_, err = svc.CreateTag(tRegForm)
		if err != nil {
			log.Printf("Error Encountered while importing tags! %v\n", err)
		}
		count++
	}

	if err != nil {
		log.Panic(err)
	}
	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("'%s' uploaded!", header.Filename),
	})
}

func (s *CCServer) ImportManyMembers(c *gin.Context) {
	instID := c.DefaultQuery("instID", "000000000000000000000000")

	file, header, err := c.Request.FormFile("file")
	defer file.Close()
	if err != nil {
		log.Panic(err)
	}

	r := csv.NewReader(file)
	count := 0
	for {
		record, err := r.Read()
		if count == 0 {
			count++
			continue
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		if len(record) < 3 {
			log.Panic("record fields are too few")
		}

		// Check if Member already exists

		phoneNum := record[len(record)-1]

		memberCount, err := svc.CountMembersByPhoneNum(phoneNum)
		if memberCount > 0 {
			log.Printf("Member PhoneNumb %v already exists!\n", phoneNum)
			continue
		}

		mRegForm := svc.MemberRegForm{
			InstID:    instID,
			FirstName: record[0],
			LastName:  record[1],
			PhoneNum:  record[len(record)-1],
		}
		_, err = svc.CreateMember(mRegForm)
		if err != nil {
			log.Printf("Error Encountered while importing members! %v\n", err)
		}
		count++
	}

	if err != nil {
		log.Panic(err)
	}
	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("'%s' uploaded!", header.Filename),
	})

}
