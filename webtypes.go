// Copyright 2020 Kevin Gentile
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

package main

//CardOption is used in templating to display options in a ConsoleCard
type CardOption struct {
	Label string `json:"label"`
	URL   string `json:"URL"`
}

//ConsoleCard is in templating to display a card on the console
type ConsoleCard struct {
	Title   string        `json:"title"`
	Options []*CardOption `json:"options"`
}
