// Copyright 2024 Buf Technologies, Inc.
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

package check

import (
	"context"
	"errors"

	"github.com/bufbuild/bufplugin-go/internal/pkg/xslices"
	"github.com/bufbuild/protovalidate-go"
)

// Spec is the spec for a plugin.
//
// It is used to construct a plugin on the server-side (i.e. within the plugin).
//
// Generally, this is provided to Main. This library will handle Check and ListRules calls
// based on the provided RuleSpecs.
type Spec struct {
	// Required.
	//
	// All RuleSpecs must have Category IDs that match a CategorySpec within Categories.
	//
	// No IDs can overlap with Category IDs in Categories.
	Rules []*RuleSpec
	// Required if any RuleSpec specifies a category.
	//
	// All CategorySpecs must have an ID that matches at least one Category ID on a
	// RuleSpec within Rules.
	//
	// No IDs can overlap with Rule IDs in Rules.
	Categories []*CategorySpec

	// Before is a function that will be executed before any RuleHandlers are
	// invoked that returns a new Context and Request. This new Context and
	// Request will be passed to the RuleHandlers. This allows for any
	// pre-processing that needs to occur.
	Before func(ctx context.Context, request Request) (context.Context, Request, error)
}

// *** PRIVATE ***

func validateSpec(validator *protovalidate.Validator, spec *Spec) error {
	if len(spec.Rules) == 0 {
		return errors.New("Spec.Rules is empty")
	}
	categoryIDs := xslices.Map(spec.Categories, func(categorySpec *CategorySpec) string { return categorySpec.ID })
	if err := validateNoDuplicateRuleOrCategoryIDs(
		append(
			xslices.Map(spec.Rules, func(ruleSpec *RuleSpec) string { return ruleSpec.ID }),
			categoryIDs...,
		),
	); err != nil {
		return err
	}
	categoryIDMap := xslices.ToStructMap(categoryIDs)
	if err := validateRuleSpecs(validator, spec.Rules, categoryIDMap); err != nil {
		return err
	}
	return validateCategorySpecs(validator, spec.Categories, spec.Rules)
}
