/*
 * Minio S3Verify Library for Amazon S3 Compatible Cloud Storage (C) 2016 Minio, Inc.
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

package main

import (
	"strings"

	"github.com/minio/cli"
	"github.com/minio/mc/pkg/console"
)

// getObjectCmd can be used to run the getobject compatibility test.
var getObjectCmd = cli.Command{
	Name:   "getobject",
	Usage:  "Run the getobject test.",
	Action: mainGetObject,
	Flags:  append(getObjectFlags, globalFlags...),
	CustomHelpTemplate: `NAME:
	s3verify {{.Name}} - {{.Usage}}

USAGE:
	s3verify {{.Name}} [COMMAND...] [FLAGS]

FLAGS:
	{{range .Flags}}{{.}}
	{{end}}

EXAMPLES:
	1. Test on the Minio server. Note that play.minio.io is a public test server. You are free to use these secret and access keys in all your tests.
		$ S3_URL=https://play.minio.io:9000 S3_ACCESS=Q3AM3UQ867SPQQA43P2F S3_SECRET=zuf+tfteSlswRu7BJ86wekitnifILbZam1KYY3TG s3verify getobject
	2. Test on the Amazon S3 server using flags. Note that passing access and secret keys as flags should be avoided on a multi-user server for security reasons.
		$ s3verify getobject --access YOUR_ACCESS_KEY --secret YOUR_SECRET_KEY --url https://s3.amazonaws.com
	`,
}

// Flags supported by the getobject command.
var (
	getObjectFlags = []cli.Flag{
		cli.BoolFlag{
			Name:  "help, h",
			Usage: "Help of get object",
		},
	}
)

// TODO: Create array of functions called GetObject. functions of type *config and returns error. Each test handles its own test.

// Messages printed during the running of the test.
const (
	noHeader          = "[3/7] GetObject (No Header):"
	rangeHeader       = "[4/7] GetObject (Range):"
	ifMatchHeader     = "[5/7] GetObject (If-Match):"
	ifNoneMatchHeader = "[6/7] GetObject (If-None-Match):"
)

// mainGetObject - Entry point for the getobject command and test.
func mainGetObject(ctx *cli.Context) {
	// TODO: Differentiate errors: s3verify vs Minio vs test failure.
	// Set up a new config.
	config := newServerConfig(ctx)
	// Test GetObject with no header set.
	if err := mainGetObjectNoHeader(*config, noHeader); err != nil {
		console.Fatalln(err)
	}
	// Erase the old progress bar
	console.Eraseline()
	// Get amount of padding needed.
	padding := messageWidth - len([]rune(noHeader))
	// Update test progress.
	console.PrintC(noHeader + strings.Repeat(" ", padding) + "[OK]\n")
	// Test GetObject with Range header set.
	if err := mainGetObjectRange(*config, rangeHeader); err != nil {
		console.Fatalln(err)
	}
	// Update amount of padding needed.
	padding = messageWidth - len([]rune(rangeHeader))
	// Erase the old progress bar
	console.Eraseline()
	// Update test progress.
	console.PrintC(rangeHeader + strings.Repeat(" ", padding) + "[OK]\n")
	// Test GetObject with If-Match header set.
	if err := mainGetObjectIfMatch(*config, ifMatchHeader); err != nil {
		console.Fatalln(err)
	}
	// Erase the old progress bar
	console.Eraseline()
	// Update amount of padding needed.
	padding = messageWidth - len([]rune(ifMatchHeader))
	// Update test progress.
	console.PrintC(ifMatchHeader + strings.Repeat(" ", padding) + "[OK]\n")
	// Test GetObject with If-None-Match header set.
	if err := mainGetObjectIfNoneMatch(*config, ifNoneMatchHeader); err != nil {
		console.Fatalln(err)
	}
	// Erase the old progress bar
	console.Eraseline()
	// Update the amound of padding needed.
	padding = messageWidth - len([]rune(ifNoneMatchHeader))
	// Update test progress.
	console.PrintC(ifNoneMatchHeader + strings.Repeat(" ", padding) + "[OK]\n")

}
