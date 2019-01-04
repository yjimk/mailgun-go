package mailgun

import (
	"strconv"
)

// A mailing list may have one of three membership modes.
// ReadOnly specifies that nobody, including Members,
// may send messages to the mailing list.
// Messages distributed on such lists come from list administrator accounts only.
// Members specifies that only those who subscribe to the mailing list may send messages.
// Everyone specifies that anyone and everyone may both read and submit messages
// to the mailing list, including non-subscribers.
const (
	ReadOnly = "readonly"
	Members  = "members"
	Everyone = "everyone"
)

// A List structure provides information for a mailing list.
//
// AccessLevel may be one of ReadOnly, Members, or Everyone.
type List struct {
	Address      string `json:"address,omitempty"`
	Name         string `json:"name,omitempty"`
	Description  string `json:"description,omitempty"`
	AccessLevel  string `json:"access_level,omitempty"`
	CreatedAt    string `json:"created_at,omitempty"`
	MembersCount int    `json:"members_count,omitempty"`
}

type ListsOptions struct {
	Limit int
}

type listsResponse struct {
	Lists  []List `json:"items"`
	Paging Paging `json:"paging"`
}

type ListsIterator struct {
	listsResponse
	mg  Mailgun
	err error
}

// ListMailingLists returns the specified set of mailing lists administered by your account.
func (mg *MailgunImpl) ListMailingLists(opts *ListsOptions) *ListsIterator {
	r := newHTTPRequest(generatePublicApiUrl(mg, listsEndpoint) + "/pages")
	r.setClient(mg.Client())
	r.setBasicAuth(basicAuthUser, mg.APIKey())
	if opts != nil {
		if opts.Limit != 0 {
			r.addParameter("limit", strconv.Itoa(opts.Limit))
		}
	}
	url, err := r.generateUrlWithParameters()
	return &ListsIterator{
		mg:            mg,
		listsResponse: listsResponse{Paging: Paging{Next: url, First: url}},
		err:           err,
	}
}

// If an error occurred during iteration `Err()` will return non nil
func (li *ListsIterator) Err() error {
	return li.err
}

// Retrieves the chronologically previous batch of events, if any exist.
// You know you're at the end of the list when len(Events())==0.
func (li *ListsIterator) GetPrevious() error {
	return li.fetch(li.Paging.Previous)
}

// Retrieves the chronologically next batch of events, if any exist.
// You know you're at the end of the list when len(Events())==0.
func (li *ListsIterator) GetNext() error {
	return li.fetch(li.Paging.Next)
}

// Retrieves the next page of events from the api. Returns false when there
// no more pages to retrieve or if there was an error. Use `.Err()` to retrieve
// the error
func (li *ListsIterator) Next(items *[]List) bool {
	if li.err != nil {
		return false
	}
	li.err = li.fetch(li.Paging.Next)
	if li.err != nil {
		return false
	}
	*items = li.Lists
	if len(li.Lists) == 0 {
		return false
	}
	return true
}

// Retrieves the first page of events from the api. Returns false if there
// was an error. It also sets the iterator object to the first page.
// Use `.Err()` to retrieve the error.
func (li *ListsIterator) First(items *[]List) bool {
	if li.err != nil {
		return false
	}
	li.err = li.fetch(li.Paging.First)
	if li.err != nil {
		return false
	}
	*items = li.Lists
	return true
}

// Retrieves the last page of events from the api.
// Calling Last() is invalid unless you first call First() or Next()
// Returns false if there was an error. It also sets the iterator object
// to the last page. Use `.Err()` to retrieve the error.
func (li *ListsIterator) Last(items *[]List) bool {
	if li.err != nil {
		return false
	}
	li.err = li.fetch(li.Paging.Last)
	if li.err != nil {
		return false
	}
	*items = li.Lists
	return true
}

// Retrieves the previous page of events from the api. Returns false when there
// no more pages to retrieve or if there was an error. Use `.Err()` to retrieve
// the error if any
func (li *ListsIterator) Previous(items *[]List) bool {
	if li.err != nil {
		return false
	}
	if li.Paging.Previous == "" {
		return false
	}
	li.err = li.fetch(li.Paging.Previous)
	if li.err != nil {
		return false
	}
	*items = li.Lists
	if len(li.Lists) == 0 {
		return false
	}
	return true
}

// GetFirstPage, GetPrevious, and GetNext all have a common body of code.
// fetch completes the API fetch common to all three of these functions.
func (li *ListsIterator) fetch(url string) error {
	r := newHTTPRequest(url)
	r.setClient(li.mg.Client())
	r.setBasicAuth(basicAuthUser, li.mg.APIKey())

	return getResponseFromJSON(r, &li.listsResponse)
}

// CreateMailingList creates a new mailing list under your Mailgun account.
// You need specify only the Address and Name members of the prototype;
// Description, and AccessLevel are optional.
// If unspecified, Description remains blank,
// while AccessLevel defaults to Everyone.
func (mg *MailgunImpl) CreateMailingList(prototype List) (List, error) {
	r := newHTTPRequest(generatePublicApiUrl(mg, listsEndpoint))
	r.setClient(mg.Client())
	r.setBasicAuth(basicAuthUser, mg.APIKey())
	p := newUrlEncodedPayload()
	if prototype.Address != "" {
		p.addValue("address", prototype.Address)
	}
	if prototype.Name != "" {
		p.addValue("name", prototype.Name)
	}
	if prototype.Description != "" {
		p.addValue("description", prototype.Description)
	}
	if prototype.AccessLevel != "" {
		p.addValue("access_level", prototype.AccessLevel)
	}
	response, err := makePostRequest(r, p)
	if err != nil {
		return List{}, err
	}
	var l List
	err = response.parseFromJSON(&l)
	return l, err
}

// DeleteMailingList removes all current members of the list, then removes the list itself.
// Attempts to send e-mail to the list will fail subsequent to this call.
func (mg *MailgunImpl) DeleteMailingList(addr string) error {
	r := newHTTPRequest(generatePublicApiUrl(mg, listsEndpoint) + "/" + addr)
	r.setClient(mg.Client())
	r.setBasicAuth(basicAuthUser, mg.APIKey())
	_, err := makeDeleteRequest(r)
	return err
}

// GetMailingList allows your application to recover the complete List structure
// representing a mailing list, so long as you have its e-mail address.
func (mg *MailgunImpl) GetMailingList(addr string) (List, error) {
	r := newHTTPRequest(generatePublicApiUrl(mg, listsEndpoint) + "/" + addr)
	r.setClient(mg.Client())
	r.setBasicAuth(basicAuthUser, mg.APIKey())
	response, err := makeGetRequest(r)
	if err != nil {
		return List{}, err
	}

	var envelope struct {
		List `json:"list"`
	}
	err = response.parseFromJSON(&envelope)
	return envelope.List, err
}

// UpdateList allows you to change various attributes of a list.
// Address, Name, Description, and AccessLevel are all optional;
// only those fields which are set in the prototype will change.
//
// Be careful!  If changing the address of a mailing list,
// e-mail sent to the old address will not succeed.
// Make sure you account for the change accordingly.
func (mg *MailgunImpl) UpdateMailingList(addr string, prototype List) (List, error) {
	r := newHTTPRequest(generatePublicApiUrl(mg, listsEndpoint) + "/" + addr)
	r.setClient(mg.Client())
	r.setBasicAuth(basicAuthUser, mg.APIKey())
	p := newUrlEncodedPayload()
	if prototype.Address != "" {
		p.addValue("address", prototype.Address)
	}
	if prototype.Name != "" {
		p.addValue("name", prototype.Name)
	}
	if prototype.Description != "" {
		p.addValue("description", prototype.Description)
	}
	if prototype.AccessLevel != "" {
		p.addValue("access_level", prototype.AccessLevel)
	}
	var l List
	response, err := makePutRequest(r, p)
	if err != nil {
		return l, err
	}
	err = response.parseFromJSON(&l)
	return l, err
}
