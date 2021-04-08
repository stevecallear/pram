package aws

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/google/uuid"
)

const (
	snsPolicyTemplateStr = `{
  "Version": "2008-10-17",
  "Id": "{{.PID}}",
  "Statement": [{
    "Sid": "{{.SID}}",
    "Effect": "Allow",
    "Principal": {
      "AWS": "*"
    },
    "Action": [
      "SNS:GetTopicAttributes",
      "SNS:SetTopicAttributes",
      "SNS:AddPermission",
      "SNS:RemovePermission",
      "SNS:DeleteTopic",
      "SNS:Subscribe",
      "SNS:ListSubscriptionsByTopic",
      "SNS:Publish",
      "SNS:Receive"
    ],
    "Resource": "{{.TopicARN}}",
    "Condition": {
      "StringEquals": {
        "AWS:SourceOwner": "{{.AccountID}}"
      }
    }
  }]
}`

	sqsPolicyTemplateStr = `{
  "Version": "2012-10-17",
  "Id": "{{.PID}}",
  "Statement": [{
    "Sid": "{{.SID}}",
    "Effect": "Allow",
    "Principal": {
      "Service": "sns.amazonaws.com"
    },
    "Action": ["sqs:SendMessage"],
    "Resource": "{{.QueueARN}}",
    "Condition": {
      "ArnEquals": {
        "AWS:SourceArn": "{{.TopicARN}}"
      }
    }
  }]
}`

	redrivePolicyTemplateStr = `{
  "deadLetterTargetArn": "{{.DeadLetterTargetARN}}",
  "maxReceiveCount": "{{.MaxReceiveCount}}"
}`
)

var (
	snsPolicyTemplate     = template.Must(template.New("snsPolicy").Parse(snsPolicyTemplateStr))
	sqsPolicyTemplate     = template.Must(template.New("sqsPolicy").Parse(sqsPolicyTemplateStr))
	redrivePolicyTemplate = template.Must(template.New("redrivePolicy").Parse(redrivePolicyTemplateStr))
)

// SNSAccessPolicy returns a new sns access policy
func SNSAccessPolicy(topicARN string) (string, error) {
	buf := bytes.NewBuffer(nil)

	aid, err := accountIDFromARN(topicARN)
	if err != nil {
		return "", err
	}

	err = snsPolicyTemplate.Execute(buf, &struct {
		PID       string
		SID       string
		TopicARN  string
		AccountID string
	}{
		PID:       strings.ReplaceAll(uuid.NewString(), "-", ""),
		SID:       strings.ReplaceAll(uuid.NewString(), "-", ""),
		TopicARN:  topicARN,
		AccountID: aid,
	})
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

// SQSAccessPolicy returns a new sqs access policy
func SQSAccessPolicy(topicARN, queueARN string) (string, error) {
	buf := bytes.NewBuffer(nil)

	err := sqsPolicyTemplate.Execute(buf, &struct {
		PID      string
		SID      string
		TopicARN string
		QueueARN string
	}{
		PID:      strings.ReplaceAll(uuid.NewString(), "-", ""),
		SID:      strings.ReplaceAll(uuid.NewString(), "-", ""),
		TopicARN: topicARN,
		QueueARN: queueARN,
	})
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

// SQSRedrivePolicy returns a new sqs redrive policy
func SQSRedrivePolicy(errorQueueARN string, maxReceiveCount int) (string, error) {
	buf := bytes.NewBuffer(nil)

	err := redrivePolicyTemplate.Execute(buf, &struct {
		DeadLetterTargetARN string
		MaxReceiveCount     int
	}{
		DeadLetterTargetARN: errorQueueARN,
		MaxReceiveCount:     maxReceiveCount,
	})
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

func accountIDFromARN(arn string) (string, error) {
	els := strings.Split(arn, ":")
	if len(els) < 5 {
		return "", fmt.Errorf("invalid arn: %s", arn)
	}

	return els[4], nil
}
