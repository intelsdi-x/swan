// Copyright (c) 2017 Intel Corporation
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

/*
Package conf extends builtin 'flag' packagea to provide:
- environment parsing with predefined prefix,
- config file generation with grouping (instead of lexicographical order),
- ability to extract current values of or registered flags (defined with wrappers),
- new types of flags e.g. SliceFlag,
- predefined flags for logging (logrus integration),
*/
package conf
