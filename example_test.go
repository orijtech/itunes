// Copyright 2018 Orijtech, Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package itunes_test

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/orijtech/itunes"
)

func ExampleClient_Search() {
	client := new(itunes.Client)
	sres, err := client.Search(context.Background(), &itunes.Search{
		Term:  "Change",
		Limit: 12,
	})

	if err != nil {
		log.Fatal(err)
	}

	for i, result := range sres.Results {
		blob, _ := json.MarshalIndent(result, "", "  ")
		fmt.Printf("#%d: result: %s\n", i, blob)
	}
}

func ExampleClient_SearchById() {
	client := new(itunes.Client)
	sres, err := client.SearchById(context.Background(), "263058648")
	if err != nil {
		log.Fatal(err)
	}

	for i, result := range sres.Results {
		blob, _ := json.MarshalIndent(result, "", "  ")
		fmt.Printf("#%d: result: %s\n", i, blob)
	}
}
