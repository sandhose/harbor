// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package daysps

import (
	"strconv"
	"testing"
	"time"

	"github.com/goharbor/harbor/src/pkg/retention/policy/rule"
	"github.com/goharbor/harbor/src/pkg/retention/res"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type EvaluatorTestSuite struct {
	suite.Suite
}

func (e *EvaluatorTestSuite) TestNew() {
	tests := []struct {
		Name      string
		args      rule.Parameters
		expectedN int
	}{
		{Name: "Valid", args: map[string]rule.Parameter{ParameterN: 5}, expectedN: 5},
		{Name: "Default If Negative", args: map[string]rule.Parameter{ParameterN: -1}, expectedN: DefaultN},
		{Name: "Default If Not Set", args: map[string]rule.Parameter{}, expectedN: DefaultN},
		{Name: "Default If Wrong Type", args: map[string]rule.Parameter{ParameterN: "foo"}, expectedN: DefaultN},
	}

	for _, tt := range tests {
		e.T().Run(tt.Name, func(t *testing.T) {
			e := New(tt.args).(*evaluator)

			require.Equal(t, tt.expectedN, e.n)
		})
	}
}

func (e *EvaluatorTestSuite) TestProcess() {
	now := time.Now().UTC()
	data := []*res.Candidate{
		{PushedTime: daysAgo(now, 1)},
		{PushedTime: daysAgo(now, 2)},
		{PushedTime: daysAgo(now, 3)},
		{PushedTime: daysAgo(now, 4)},
		{PushedTime: daysAgo(now, 5)},
		{PushedTime: daysAgo(now, 10)},
		{PushedTime: daysAgo(now, 20)},
		{PushedTime: daysAgo(now, 30)},
	}

	tests := []struct {
		n           int
		expected    int
		minPushTime int64
	}{
		{n: 0, expected: 0, minPushTime: 0},
		{n: 1, expected: 1, minPushTime: daysAgo(now, 1)},
		{n: 2, expected: 2, minPushTime: daysAgo(now, 2)},
		{n: 3, expected: 3, minPushTime: daysAgo(now, 3)},
		{n: 4, expected: 4, minPushTime: daysAgo(now, 4)},
		{n: 5, expected: 5, minPushTime: daysAgo(now, 5)},
		{n: 15, expected: 6, minPushTime: daysAgo(now, 10)},
		{n: 90, expected: 8, minPushTime: daysAgo(now, 30)},
	}

	for _, tt := range tests {
		e.T().Run(strconv.Itoa(tt.n), func(t *testing.T) {
			sut := New(map[string]rule.Parameter{ParameterN: tt.n})

			result, err := sut.Process(data)

			require.NoError(t, err)
			require.Len(t, result, tt.expected)

			for _, v := range result {
				assert.False(t, v.PushedTime < tt.minPushTime)
			}
		})
	}
}

func TestEvaluatorSuite(t *testing.T) {
	suite.Run(t, &EvaluatorTestSuite{})
}

func daysAgo(from time.Time, n int) int64 {
	return from.Add(time.Duration(-1*24*n) * time.Hour).Unix()
}
