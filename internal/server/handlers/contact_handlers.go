package handlers

import (
	"net/http"

	"github.com/dockworks/dm-web-backend/internal/db"
	"github.com/dockworks/dm-web-backend/internal/requests"
	"github.com/dockworks/dm-web-backend/internal/responses"
	s "github.com/dockworks/dm-web-backend/internal/server"
	"github.com/dockworks/dm-web-backend/pkg/utils"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// ContactHandler handles contact-related operations
type ContactHandler struct {
	server *s.Server
}

// NewContactHandler creates a new contact handler
func NewContactHandler(server *s.Server) *ContactHandler {
	return &ContactHandler{server: server}
}

// ListContacts lists all contacts for a marina
//
//	@Summary		List contacts
//	@Description	Lists all contacts for a marina, optionally filtered by type
//	@Tags			Contacts
//	@Accept			json
//	@Produce		json
//	@Param			marinaId	path		string	true	"Marina ID"	Format(uuid)
//	@Param			type		query		string	false	"Contact type (phone or email)"
//	@Success		200			{object}	responses.ContactListResponse
//	@Failure		400			{object}	responses.Error
//	@Failure		500			{object}	responses.Error
//	@Security		ApiKeyAuth
//	@Router			/marina/{marinaId}/contacts [get]
func (h *ContactHandler) ListContacts(c echo.Context) error {
	marinaIDStr := c.Param("id")
	marinaID, err := uuid.Parse(marinaIDStr)
	if err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, "Invalid marina ID").JSON(c)
	}

	contactType := c.QueryParam("type")
	var contacts []db.Contact
	var err2 error

	if contactType != "" {
		contacts, err2 = h.server.DB.Queries().ListContactsByType(c.Request().Context(), db.ListContactsByTypeParams{
			MarinaID: marinaID,
			Type:     contactType,
		})
	} else {
		contacts, err2 = h.server.DB.Queries().ListContacts(c.Request().Context(), marinaID)
	}

	if err2 != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, err2).JSON(c)
	}

	return responses.NewContactListResponse(contacts).JSON(c)
}

// CreateContact creates a new contact
//
//	@Summary		Create contact
//	@Description	Creates a new contact for a marina
//	@Tags			Contacts
//	@Accept			json
//	@Produce		json
//	@Param			marinaId	path		string						true	"Marina ID"	Format(uuid)
//	@Param			contact		body		requests.CreateContactRequest	true	"Contact details"
//	@Success		201			{object}	responses.ContactResponse
//	@Failure		400			{object}	responses.Error
//	@Failure		500			{object}	responses.Error
//	@Security		ApiKeyAuth
//	@Router			/marina/{marinaId}/contacts [post]
func (h *ContactHandler) CreateContact(c echo.Context) error {
	marinaIDStr := c.Param("id")
	marinaID, err := uuid.Parse(marinaIDStr)
	if err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, "Invalid marina ID").JSON(c)
	}

	var req requests.CreateContactRequest
	if err := c.Bind(&req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	if err := c.Validate(&req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	// Convert string fields to pointers for optional fields
	contactType := string(req.Type)
	var description, email, phone *string
	if req.Description != "" {
		description = &req.Description
	}
	if req.Email != "" {
		email = &req.Email
	}
	if req.Phone != "" {
		phone = &req.Phone
	}
	if contactType == "phone" && phone == nil {
		return responses.NewErrorResponse(http.StatusBadRequest, "Phone is required for phone contacts").JSON(c)
	}
	if contactType == "email" && email == nil {
		return responses.NewErrorResponse(http.StatusBadRequest, "Email is required for email contacts").JSON(c)
	}
	// remove phone if contact type is email
	if contactType == "email" {
		phone = nil
	}
	// remove email if contact type is phone
	if contactType == "phone" {
		email = nil
	}
	var isCpContact *bool
	if req.IsCPContact != nil {
		isCpContact = req.IsCPContact
	} else {
		isCpContact = utils.Pointer(false)
	}
	if isCpContact != nil && *isCpContact {
		// unset all cp contacts
		err = h.server.DB.Queries().UnsetCPContact(c.Request().Context(), marinaID)
		if err != nil {
			return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
		}
	}
	params := db.CreateContactParams{
		MarinaID:    marinaID,
		Type:        contactType,
		Name:        req.Name,
		Description: description,
		Email:       email,
		Phone:       phone,
		IsCpContact: isCpContact,
	}

	contact, err := h.server.DB.Queries().CreateContact(c.Request().Context(), params)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	return responses.NewContactResponse(contact).JSON(c)
}

// UpdateContact updates an existing contact
//
//	@Summary		Update contact
//	@Description	Updates an existing contact
//	@Tags			Contacts
//	@Accept			json
//	@Produce		json
//	@Param			marinaId	path		string						true	"Marina ID"	Format(uuid)
//	@Param			contactId	path		string						true	"Contact ID"	Format(uuid)
//	@Param			contact		body		requests.UpdateContactRequest	true	"Contact details"
//	@Success		200			{object}	responses.ContactResponse
//	@Failure		400			{object}	responses.Error
//	@Failure		404			{object}	responses.Error
//	@Failure		500			{object}	responses.Error
//	@Security		ApiKeyAuth
//	@Router			/marina/{marinaId}/contacts/{contactId} [put]
func (h *ContactHandler) UpdateContact(c echo.Context) error {
	contactIDStr := c.Param("contactId")
	contactID, err := uuid.Parse(contactIDStr)
	if err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, "Invalid contact ID").JSON(c)
	}

	var req requests.UpdateContactRequest
	if err := c.Bind(&req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	if err := c.Validate(&req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	contact, err := h.server.DB.Queries().GetContactByID(c.Request().Context(), contactID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusNotFound, "Contact not found").JSON(c)
	}

	// Convert string fields to pointers for optional fields
	var name, description, email, phone *string
	if req.Description != "" {
		description = &req.Description
	} else {
		description = contact.Description
	}
	if req.Name != "" {
		name = &req.Name
	} else {
		name = &contact.Name
	}
	if req.Email != "" {
		email = &req.Email
	} else {
		email = contact.Email
	}
	if req.Phone != "" {
		phone = &req.Phone
	} else {
		phone = contact.Phone
	}

	if req.Type == "phone" && phone == nil {
		return responses.NewErrorResponse(http.StatusBadRequest, "Phone is required for phone contacts").JSON(c)
	}
	if req.Type == "email" && email == nil {
		return responses.NewErrorResponse(http.StatusBadRequest, "Email is required for email contacts").JSON(c)
	}

	// remove phone if contact type is email
	if req.Type == "email" {
		phone = nil
	}
	// remove email if contact type is phone
	if req.Type == "phone" {
		email = nil
	}
	// Check if isCPContact is provided in the request and if it is true, unset all cp contacts,
	var isCpContact *bool
	if req.IsCPContact != nil {
		isCpContact = req.IsCPContact
	} else {
		isCpContact = contact.IsCpContact
	}
	if isCpContact != nil && *isCpContact {
		// unset all cp contacts
		err = h.server.DB.Queries().UnsetCPContact(c.Request().Context(), contact.MarinaID)
		if err != nil {
			return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
		}
	}
	params := db.UpdateContactParams{
		ID:          contactID,
		Type:        string(req.Type),
		Name:        *name,
		Description: description,
		Email:       email,
		Phone:       phone,
		IsCpContact: isCpContact,
	}

	newContact, err := h.server.DB.Queries().UpdateContact(c.Request().Context(), params)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	return responses.NewContactResponse(newContact).JSON(c)
}

// DeleteContact deletes a contact
//
//	@Summary		Delete contact
//	@Description	Soft deletes a contact
//	@Tags			Contacts
//	@Accept			json
//	@Produce		json
//	@Param			marinaId	path		string	true	"Marina ID"	Format(uuid)
//	@Param			contactId	path		string	true	"Contact ID"	Format(uuid)
//	@Success		200			{object}	responses.BaseResponse
//	@Failure		400			{object}	responses.Error
//	@Failure		500			{object}	responses.Error
//	@Security		ApiKeyAuth
//	@Router			/marina/{marinaId}/contacts/{contactId} [delete]
func (h *ContactHandler) DeleteContact(c echo.Context) error {
	contactIDStr := c.Param("contactId")
	contactID, err := uuid.Parse(contactIDStr)
	if err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, "Invalid contact ID").JSON(c)
	}

	err = h.server.DB.Queries().DeleteContact(c.Request().Context(), contactID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	return responses.NewSuccessResponse("Contact deleted successfully").JSON(c)
}
