package api

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type DonationListOptions struct {
	Page    int
	PerPage int
	Name    string
}

func (c *Client) ListDonations(ctx context.Context, options DonationListOptions) (*DonationsResponse, error) {
	query := url.Values{}
	if options.Page > 0 {
		query.Set("page", strconv.Itoa(options.Page))
	}
	if options.PerPage > 0 {
		query.Set("per_page", strconv.Itoa(options.PerPage))
	}
	if value := strings.TrimSpace(options.Name); value != "" {
		query.Set("name", value)
	}

	req, err := c.newRequest(ctx, http.MethodGet, "/donations", query, nil)
	if err != nil {
		return nil, err
	}

	var response DonationsResponse
	if err := c.do(req, &response); err != nil {
		return nil, err
	}

	return &response, nil
}

func (c *Client) DonationTotalByUsername(ctx context.Context, username string, maxPages int) (float64, string, error) {
	username = strings.TrimSpace(username)
	if username == "" {
		return 0, "", nil
	}
	if maxPages <= 0 {
		maxPages = 3
	}

	total := 0.0
	currency := ""

	for page := 1; page <= maxPages; page++ {
		response, err := c.ListDonations(ctx, DonationListOptions{
			Page:    page,
			PerPage: 100,
			Name:    username,
		})
		if err != nil {
			return 0, "", err
		}
		if response == nil || len(response.Data) == 0 {
			break
		}

		for _, donation := range response.Data {
			if !strings.EqualFold(strings.TrimSpace(donation.Name), username) && !strings.EqualFold(strings.TrimSpace(donation.Identifier), username) {
				continue
			}
			total += donation.Amount
			if currency == "" {
				currency = strings.TrimSpace(donation.Currency)
			}
		}

		if response.NextPageURL == "" || response.CurrentPage >= response.LastPage {
			break
		}
	}

	return total, currency, nil
}
