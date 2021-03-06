/*
 * s3verify (C) 2016 Minio, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package cmd

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"mime/multipart"
	"net/http"
	"time"

	"github.com/minio/s3verify/signv4"
)

const iso8601DateFormat = "20060102T150405Z"

// newPostObjectReq - create a new postObject Request.
func newPostObjectReq(config ServerConfig, bucketName, objectName string, objectData []byte) (*http.Request, error) {
	targetURL, err := makeTargetURL(config.Endpoint, bucketName, "", config.Region, nil)
	if err != nil {
		return nil, err
	}
	// Keep time.
	t := time.Now().UTC()
	// Expire the request five minutes from now.
	expirationTime := t.Add(time.Minute * 5)
	// Get the user credential.
	credential := signv4.GetCredential(config.Access, config.Region, t)
	// Create a new post policy.
	policy := newPostPolicyBytes(credential, bucketName, objectName, expirationTime)
	encodedPolicy := base64.StdEncoding.EncodeToString(policy)

	// Presign with V4 signature based on the policy.
	signature := signv4.PostPresignSignatureV4(encodedPolicy, t, config.Secret, config.Region)

	formData := map[string]string{
		"bucket":           bucketName,
		"key":              objectName,
		"x-amz-credential": credential,
		"policy":           encodedPolicy,
		"x-amz-signature":  signature,
		"x-amz-date":       t.Format(iso8601DateFormat),
		"x-amz-algorithm":  "AWS4-HMAC-SHA256",
	}

	// Create the multipart form.
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)

	// Set the normal formData
	for k, v := range formData {
		w.WriteField(k, v)
	}
	// Set the File formData
	writer, err := w.CreateFormFile("file", "s3verify/post/object")
	if err != nil {
		return nil, err
	}
	writer.Write(objectData)
	// Close before creating the new request.
	w.Close()

	reader := bytes.NewReader(buf.Bytes())
	// Create a new POST HTTP request.
	req, err := http.NewRequest("POST", targetURL.String(), reader)
	if err != nil {
		return nil, err
	}

	// Set the headers.
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.Header.Set("User-Agent", appUserAgent)

	return req, nil

}

// postObjectVerify - verify the response returned.
func postObjectVerify(res *http.Response, expectedStatusCode int) error {
	if err := verifyHeaderPostObject(res.Header); err != nil {
		return err
	}
	if err := verifyStatusPostObject(res.StatusCode, expectedStatusCode); err != nil {
		return err
	}
	if err := verifyBodyPostObject(res.Body); err != nil {
		return err
	}
	return nil
}

// verifyStatusPostObject - verify the status returned is the same.
func verifyStatusPostObject(respStatusCode, expectedStatusCode int) error {
	if respStatusCode != expectedStatusCode {
		err := fmt.Errorf("Unexpected Response Status Code: wanted %d, got %d", expectedStatusCode, respStatusCode)
		return err
	}
	return nil
}

// verifyHeaderPostObject - verify the header returned matches what is expected.
func verifyHeaderPostObject(header http.Header) error {
	if err := verifyStandardHeaders(header); err != nil {
		return err
	}
	return nil
}

// verifyBodyPostObject - verify that the body returned is empty.
func verifyBodyPostObject(resBody io.Reader) error {
	body, err := ioutil.ReadAll(resBody)
	if err != nil {
		return err
	}
	// A Post Object request should give back an empty body.
	if !bytes.Equal(body, []byte{}) {
		err := fmt.Errorf("Unexpected Body Received: expected empty body but received: %s", string(body))
		return err
	}
	return nil
}

// mainPostObject - entry point for the postobject test.
func mainPostObject(config ServerConfig, curTest int) bool {
	message := fmt.Sprintf("[%02d/%d] PostObject:", curTest, globalTotalNumTest)
	// Spin scanBar
	scanBar(message)

	// Post the object to the s3verify created bucket.
	bucketName := s3verifyBuckets[0].Name
	testObject := &ObjectInfo{
		Key:  "s3verify/put/object/post",
		Body: []byte(randString(60, rand.NewSource(time.Now().UnixNano()), "s3verify post data: ")),
	}

	// Spin scanBar
	scanBar(message)
	// Create a new request.
	req, err := newPostObjectReq(config, bucketName, testObject.Key, testObject.Body)
	if err != nil {
		printMessage(message, err)
		return false
	}

	// Execute the request.
	res, err := config.Client.Do(req)
	if err != nil {
		printMessage(message, err)
		return false
	}
	defer closeResponse(res)
	// Spin scanBar
	scanBar(message)
	// Verify the response.
	if err := postObjectVerify(res, http.StatusNoContent); err != nil {
		printMessage(message, err)
		return false
	}
	// Store this object in the global list of objects.
	s3verifyObjects = append(s3verifyObjects, testObject)
	// Spin scanBar
	scanBar(message)
	// Test passed.
	printMessage(message, nil)
	return true
}
