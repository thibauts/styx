// Copyright 2021 Dataptive SAS.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package client

//
type LogInfo struct {
	Name          string           `json:"name"`
	Status        string           `json:"status"`
	RecordCount   int64            `json:"record_count"`
	FileSize      int64            `json:"file_size"`
	StartPosition int64            `json:"start_position"`
	EndPosition   int64            `json:"end_position"`
}

//
type LogConfig struct {
	MaxRecordSize   int   `schema:"max_record_size"`
	IndexAfterSize  int64 `schema:"index_after_size"`
	SegmentMaxCount int64 `schema:"segment_max_count"`
	SegmentMaxSize  int64 `schema:"segment_max_size"`
	SegmentMaxAge   int64 `schema:"segment_max_age"`
	LogMaxCount     int64 `schema:"log_max_count"`
	LogMaxSize      int64 `schema:"log_max_size"`
	LogMaxAge       int64 `schema:"log_max_age"`
}

type createLogForm struct {
	Name string `schema:"name,required"`
	*LogConfig
}

//
type RestoreLogParams struct {
	Name string `schema:"name,required"`
}

//
type ListLogsResponse []LogInfo

//
type CreateLogResponse LogInfo

//
type GetLogResponse LogInfo

// //
// type ProduceResponse struct {
// 	Position int64 `json:"position"`
// 	Count    int64 `json:"count"`
// }

// //
// type ConsumeParams struct {
// 	Whence   string     `schema:"whence"`
// 	Position int64      `schema:"position"`
// 	Count    int64      `schema:"count"`
// 	Follow   bool       `schema:"follow"`
// }
