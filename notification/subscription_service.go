package notification

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/philips-software/go-hsdp-api/internal"
)

type SubscriptionService struct {
	client *Client

	validate *validator.Validate
}

type Subscription struct {
	ID                   string `json:"_id,omitempty"`
	ResourceType         string `json:"resourceType,omitempty"`
	TopicID              string `json:"topicId" validate:"required"`
	SubscriberID         string `json:"subscriberId" validate:"required"`
	SubscriptionEndpoint string `json:"subscriptionEndpoint" validate:"required"`
}

func (p *SubscriptionService) CreateSubscription(subscription Subscription) (*Subscription, *Response, error) {
	if err := p.validate.Struct(subscription); err != nil {
		return nil, nil, err
	}
	req, err := p.client.newNotificationRequest("POST", "core/notification/Subscription", subscription, nil)
	if err != nil {
		return nil, nil, err
	}
	var createdSubscription Subscription
	resp, err := p.client.do(req, &createdSubscription)
	if (err != nil && err != io.EOF) || resp == nil {
		if resp == nil && err != nil {
			err = fmt.Errorf("CreateSubscription: %w", ErrEmptyResult)
		}
		return nil, resp, err
	}
	return &createdSubscription, resp, nil
}

func (p *SubscriptionService) GetSubscriptions(opt *GetOptions, options ...OptionFunc) ([]*Subscription, *Response, error) {
	var subscriptions []*Subscription

	req, err := p.client.newNotificationRequest("GET", "core/notification/Subscription", opt, options...)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Api-Version", APIVersion)

	var bundleResponse internal.Bundle

	resp, err := p.client.do(req, &bundleResponse)
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			return nil, resp, ErrEmptyResult
		}
		return nil, resp, err
	}
	if bundleResponse.Total == 0 {
		return subscriptions, resp, ErrEmptyResult
	}
	for _, e := range bundleResponse.Entry {
		c := new(Subscription)
		if err := json.Unmarshal(e.Resource, c); err == nil {
			subscriptions = append(subscriptions, c)
		} else {
			return nil, resp, err
		}
	}
	return subscriptions, resp, err
}

func (p *SubscriptionService) DeleteSubscription(subscription Subscription) (bool, *Response, error) {
	req, err := p.client.newNotificationRequest("DELETE", "core/notification/Subscription/"+subscription.ID, nil, nil)
	if err != nil {
		return false, nil, err
	}
	req.Header.Set("api-version", APIVersion)

	var deleteResponse bytes.Buffer

	resp, err := p.client.do(req, &deleteResponse)
	if resp == nil || resp.StatusCode != http.StatusNoContent {
		return false, resp, nil
	}
	return true, resp, err
}