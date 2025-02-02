package notification

import (
	"fmt"
	"io"
	"net/http"

	"github.com/cenkalti/backoff/v4"
	"github.com/go-playground/validator/v10"
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
	SubscriptionARN      string `json:"subscriptionArn,omitempty"`
}

type ConfirmRequest struct {
	AuthenticateOnUnsubscribe string `json:"authenticateOnUnsubscribe,omitempty"`
	Token                     string `json:"token" validate:"required"`
	TopicARN                  string `json:"topicArn" validate:"required"`
	Endpoint                  string `json:"endpoint" validate:"required"`
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

func (p *SubscriptionService) GetSubscriptions(opt *GetOptions, options ...OptionFunc) ([]Subscription, *Response, error) {
	var subscriptions []Subscription

	req, err := p.client.newNotificationRequest("GET", "core/notification/Subscription", opt, options...)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Api-Version", APIVersion)

	var bundleResponse struct {
		ResourceType string         `json:"resourceType,omitempty"`
		Type         string         `json:"type,omitempty"`
		Total        int            `json:"total"`
		Entry        []Subscription `json:"entry"`
	}
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
		subscriptions = append(subscriptions, e)
	}
	return subscriptions, resp, err
}

func (p *SubscriptionService) GetSubscription(id string) (*Subscription, *Response, error) {
	subscriptions, resp, err := p.GetSubscriptions(&GetOptions{ID: &id})
	if err != nil {
		return nil, resp, err
	}
	if subscriptions == nil || len(subscriptions) != 1 {
		return nil, resp, fmt.Errorf("GetSubscriber: not found")
	}
	return &subscriptions[0], resp, nil
}

func (p *SubscriptionService) DeleteSubscription(subscription Subscription) (bool, *Response, error) {
	req, err := p.client.newNotificationRequest("DELETE", "core/notification/Subscription/"+subscription.ID, nil, nil)
	if err != nil {
		return false, nil, err
	}
	req.Header.Set("api-version", APIVersion)

	resp, err := p.client.do(req, nil)
	if err != nil {
		return false, resp, err
	}
	if resp != nil && resp.StatusCode != http.StatusNoContent {
		return false, resp, fmt.Errorf("DeleteSubscription: HTTP %d", resp.StatusCode)
	}
	return true, resp, err
}

func (p *SubscriptionService) ConfirmSubscription(confirm ConfirmRequest) (*Subscription, *Response, error) {
	var confirmResponse Subscription
	var resp *Response

	if err := p.validate.Struct(confirm); err != nil {
		return nil, nil, err
	}
	operation := func() error {
		req, err := p.client.newNotificationRequest("POST", "core/notification/Subscription/_confirm", confirm, nil)
		if err != nil {
			return err
		}
		req.Header.Set("api-version", APIVersion)

		resp, err = p.client.do(req, &confirmResponse)
		if err != nil {
			return err
		}
		if resp != nil && resp.StatusCode != http.StatusCreated {
			return fmt.Errorf("ConfirmSubscription: HTTP %d", resp.StatusCode)
		}
		return err
	}
	err := backoff.Retry(operation, backoff.WithMaxRetries(backoff.NewExponentialBackOff(), 8))
	if err != nil {
		return nil, resp, err
	}
	return &confirmResponse, resp, nil
}
