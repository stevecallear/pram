package aws_test

import (
	"encoding/json"
	"testing"

	"github.com/tidwall/gjson"

	"github.com/stevecallear/pram/internal/assert"
	"github.com/stevecallear/pram/internal/aws"
)

func TestSNSAccessPolicy(t *testing.T) {
	const arn = "arn:aws:sns:eu-west-1:111122223333:stage-package-Message"

	t.Run("should return an error if the arn is invalid", func(t *testing.T) {
		_, err := aws.SNSAccessPolicy("invalid")
		assert.ErrorExists(t, err, true)
	})

	t.Run("should generate valid json", func(t *testing.T) {
		p, err := aws.SNSAccessPolicy(arn)
		assert.ErrorExists(t, err, false)

		err = json.Unmarshal([]byte(p), &map[string]interface{}{})
		assert.ErrorExists(t, err, false)
	})

	t.Run("should return the policy", func(t *testing.T) {
		p, err := aws.SNSAccessPolicy(arn)
		assert.ErrorExists(t, err, false)

		if act, exp := gjson.Get(p, "Statement.0.Resource").Str, arn; act != exp {
			t.Errorf("got %s, expected %s", act, exp)
		}

		if act, exp := gjson.Get(p, "Statement.0.Condition.StringEquals.AWS:SourceOwner").Str, "111122223333"; act != exp {
			t.Errorf("got %s, expected %s", act, exp)
		}
	})
}

func TestSQSAccessPolicy(t *testing.T) {
	const topicARN = "arn:aws:sns:eu-west-1:111122223333:stage-package-Message"
	const queueARN = "arn:aws:sqs:eu-west-1:111122223333:stage-service-package-Message"

	t.Run("should generate valid json", func(t *testing.T) {
		p, err := aws.SQSAccessPolicy(topicARN, queueARN)
		assert.ErrorExists(t, err, false)

		err = json.Unmarshal([]byte(p), &map[string]interface{}{})
		assert.ErrorExists(t, err, false)
	})

	t.Run("should return the policy", func(t *testing.T) {
		p, err := aws.SQSAccessPolicy(topicARN, queueARN)
		assert.ErrorExists(t, err, false)

		if act, exp := gjson.Get(p, "Statement.0.Resource").Str, queueARN; act != exp {
			t.Errorf("got %s, expected %s", act, exp)
		}

		if act, exp := gjson.Get(p, "Statement.0.Condition.ArnEquals.AWS:SourceArn").Str, topicARN; act != exp {
			t.Errorf("got %s, expected %s", act, exp)
		}
	})
}

func TestSQSRedrivePolicy(t *testing.T) {
	const errorQueueARN = "arn:aws:sqs:eu-west-1:111122223333:stage-service-package-Message"
	const maxReceiveCount = 5

	t.Run("should generate valid json", func(t *testing.T) {
		p, err := aws.SQSRedrivePolicy(errorQueueARN, maxReceiveCount)
		assert.ErrorExists(t, err, false)

		err = json.Unmarshal([]byte(p), &map[string]interface{}{})
		assert.ErrorExists(t, err, false)
	})

	t.Run("should return the policy", func(t *testing.T) {
		p, err := aws.SQSRedrivePolicy(errorQueueARN, maxReceiveCount)
		assert.ErrorExists(t, err, false)

		if act, exp := gjson.Get(p, "deadLetterTargetArn").Str, errorQueueARN; act != exp {
			t.Errorf("got %s, expected %s", act, exp)
		}

		if act, exp := gjson.Get(p, "maxReceiveCount").Int(), int64(maxReceiveCount); act != exp {
			t.Errorf("got %d, expected %d", act, exp)
		}
	})
}
