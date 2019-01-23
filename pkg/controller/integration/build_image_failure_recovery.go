/*
Licensed to the Apache Software Foundation (ASF) under one or more
contributor license agreements.  See the NOTICE file distributed with
this work for additional information regarding copyright ownership.
The ASF licenses this file to You under the Apache License, Version 2.0
(the "License"); you may not use this file except in compliance with
the License.  You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package integration

import (
	"context"
	"time"

	"github.com/jpillora/backoff"

	"github.com/apache/camel-k/pkg/apis/camel/v1alpha1"
	"github.com/apache/camel-k/pkg/platform"
	"github.com/sirupsen/logrus"
)

// NewErrorRecoveryAction creates a new error recovering handling action for the integration
func NewErrorRecoveryAction() Action {
	//TODO: externalize options
	return &errorRecoveryAction{
		backOff: backoff.Backoff{
			Min:    5 * time.Second,
			Max:    1 * time.Minute,
			Factor: 2,
			Jitter: false,
		},
	}
}

type errorRecoveryAction struct {
	baseAction
	backOff backoff.Backoff
}

func (action *errorRecoveryAction) Name() string {
	return "error-recovery"
}

func (action *errorRecoveryAction) CanHandle(integration *v1alpha1.Integration) bool {
	return integration.Status.Phase == v1alpha1.IntegrationPhaseBuildFailureRecovery
}

func (action *errorRecoveryAction) Handle(ctx context.Context, integration *v1alpha1.Integration) error {
	// The integration platform needs to be initialized before starting handle
	// context in status error
	if _, err := platform.GetCurrentPlatform(ctx, action.client, integration.Namespace); err != nil {
		logrus.Info("Waiting for a integration platform to be initialized")
		return nil
	}

	if integration.Status.Failure.Recovery.Attempt > integration.Status.Failure.Recovery.AttemptMax {
		target := integration.DeepCopy()
		target.Status.Phase = v1alpha1.IntegrationPhaseError

		logrus.Infof("Max recovery attempt reached for integration %s, transition to phase %s", integration.Name, string(target.Status.Phase))

		return action.client.Status().Update(ctx, target)
	}

	if integration.Status.Failure != nil {
		lastAttempt := integration.Status.Failure.Recovery.AttemptTime
		if lastAttempt.IsZero() {
			lastAttempt = integration.Status.Failure.Time
		}

		elapsed := time.Since(lastAttempt).Seconds()
		elapsedMin := action.backOff.ForAttempt(float64(integration.Status.Failure.Recovery.Attempt)).Seconds()

		if elapsed < elapsedMin {
			return nil
		}

		target := integration.DeepCopy()
		target.Status = v1alpha1.IntegrationStatus{}
		target.Status.Phase = ""
		target.Status.Failure = integration.Status.Failure
		target.Status.Failure.Recovery.Attempt = integration.Status.Failure.Recovery.Attempt + 1
		target.Status.Failure.Recovery.AttemptTime = time.Now()

		logrus.Infof("Recovery attempt for integration %s (%d/%d)",
			integration.Name,
			target.Status.Failure.Recovery.Attempt,
			target.Status.Failure.Recovery.AttemptMax,
		)

		return action.client.Status().Update(ctx, target)
	}

	return nil
}