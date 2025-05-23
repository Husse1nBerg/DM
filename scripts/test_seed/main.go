package main

import (
	"context"
	"log"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/dockworks/dm-web-backend/internal/config"
	sqlc "github.com/dockworks/dm-web-backend/internal/db"
	conn "github.com/dockworks/dm-web-backend/internal/pg"
	"github.com/dockworks/dm-web-backend/pkg/models"
	u "github.com/dockworks/dm-web-backend/pkg/utils"
)

// RunTestSeed seeds the test database with initial data for testing
func main() {
	cfg := config.New()
	ctx := context.Background()
	db := conn.NewConnection(&cfg.TestDB)
	q := db.Queries()

	// Test Organization
	testOrg, err := q.GetOrganizationByEmail(ctx, "test@marinamanager.com")
	if err != nil {
		if err == pgx.ErrNoRows {
			orgAddress, err := q.CreateAddress(ctx, sqlc.CreateAddressParams{
				Street:     u.Pointer("123 Test Street"),
				City:       u.Pointer("Test City"),
				State:      u.Pointer("Test State"),
				PostalCode: u.Pointer("12345"),
				Country:    u.Pointer("Test Country"),
				Latitude:   35.1234,
				Longitude:  -120.4321,
			})
			if err != nil {
				log.Fatalf("failed to create org address: %v", err)
			}

			testOrg, err = q.CreateOrganization(ctx, sqlc.CreateOrganizationParams{
				Email:     "test@marinamanager.com",
				Name:      "Test Marina Organization",
				Image:     u.Pointer("https://example.com/test-org-logo.png"),
				Website:   u.Pointer("https://www.testmarinaorg.com"),
				Country:   u.Pointer("Test Country"),
				Phone:     u.Pointer("+1234567890"),
				IsActive:  u.Pointer(true),
				IsTest:    u.Pointer(true),
				AddressID: orgAddress.ID,
			})
			if err != nil {
				log.Fatalf("failed to create test organization: %v", err)
			}
			log.Println("Created test organization: Test Marina Organization")
		} else {
			log.Fatalf("failed to get organization: %v", err)
		}
	} else {
		log.Println("Test organization already exists")
	}

	// Test Marinas
	marinaNames := []string{"Test Marina Alpha", "Test Marina Beta"}
	marinaEmails := []string{"marina-alpha@test.com", "marina-beta@test.com"}

	var testMarinas []sqlc.Marina

	for i, name := range marinaNames {
		marina, err := q.GetMarinaByEmail(ctx, marinaEmails[i])
		if err != nil {
			if err == pgx.ErrNoRows {
				marinaAddress, err := q.CreateAddress(ctx, sqlc.CreateAddressParams{
					Street:     u.Pointer(string(rune(65+i)) + " Marina Drive"),
					City:       u.Pointer("Marina City"),
					State:      u.Pointer("Marina State"),
					PostalCode: u.Pointer("54321"),
					Country:    u.Pointer("Marina Country"),
					Latitude:   34.1234 + float64(i),
					Longitude:  -119.4321 - float64(i),
				})
				if err != nil {
					log.Fatalf("failed to create marina address: %v", err)
				}

				workingHours := models.DefaultWorkingHours()
				workingHoursBytes, err := workingHours.ToBytes()
				if err != nil {
					log.Fatalf("failed to create working hours: %v", err)
				}

				marina, err = q.CreateMarina(ctx, sqlc.CreateMarinaParams{
					Name:           name,
					Email:          marinaEmails[i],
					Location:       u.Pointer("Test Location " + string(rune(65+i))),
					Phone:          u.Pointer("+1987654" + string(rune(48+i))),
					Country:        u.Pointer("Marina Country"),
					Currency:       u.Pointer("USD"),
					WorkingHours:   workingHoursBytes,
					Website:        u.Pointer("https://www." + name + ".com"),
					Image:          u.Pointer("https://example.com/" + name + "-logo.png"),
					MaxUsers:       u.Pointer(int32(100)),
					IsActive:       u.Pointer(true),
					IsTest:         u.Pointer(true),
					OrganizationID: testOrg.ID,
					AddressID:      marinaAddress.ID,
				})
				if err != nil {
					log.Fatalf("failed to create marina: %v", err)
				}
				testMarinas = append(testMarinas, marina)
				log.Printf("Created test marina: %s", name)
			} else {
				log.Fatalf("failed to get marina: %v", err)
			}
		} else {
			testMarinas = append(testMarinas, marina)
			log.Printf("Test marina already exists: %s", name)
		}
	}

	// Test Roles
	roleNames := []string{
		"test_superuser",
		"test_org_admin",
		"test_marina_admin",
		"test_staff",
		"test_viewer",
	}

	roleDescriptions := []string{
		"Test Superuser with full system access",
		"Test Organization Administrator",
		"Test Administrator at the marina level",
		"Test Regular staff member with limited write access",
		"Test Read-only access",
	}

	rolePermissions := []string{
		`{"all": true, "admin_panel": true, "manage_org": true, "manage_marina": true, "manage_users": true, "manage_roles": true}`,
		`{"manage_org": true, "manage_marina": true, "manage_users": true, "read": true, "create": true, "update": true, "delete": true}`,
		`{"manage_marina": true, "manage_users": true, "read": true, "create": true, "update": true, "delete": true}`,
		`{"read": true, "create": true, "update": true}`,
		`{"read": true}`,
	}

	var testRoles []sqlc.Role

	for i, name := range roleNames {
		role, err := q.GetRoleByName(ctx, name)
		if err != nil {
			if err == pgx.ErrNoRows {
				role, err = q.CreateRole(ctx, sqlc.CreateRoleParams{
					Name:        name,
					Description: u.Pointer(roleDescriptions[i]),
					Permissions: []byte(rolePermissions[i]),
					IsActive:    u.Pointer(true),
				})
				if err != nil {
					log.Fatalf("failed to create role %s: %v", name, err)
				}
				testRoles = append(testRoles, role)
				log.Printf("Created test role: %s", name)
			} else {
				log.Fatalf("failed to get role: %v", err)
			}
		} else {
			testRoles = append(testRoles, role)
			log.Printf("Test role already exists: %s", name)
		}
	}

	// Test Users
	usernames := []string{"test_super", "test_orgadmin", "test_marina1admin", "test_marina2admin", "test_staff1", "test_staff2", "test_viewer"}
	firstNames := []string{"Super", "Org", "Marina1", "Marina2", "Staff", "Staff", "View"}
	lastNames := []string{"User", "Admin", "Admin", "Admin", "User1", "User2", "User"}
	emails := []string{
		"super@test.com",
		"orgadmin@test.com",
		"marina1admin@test.com",
		"marina2admin@test.com",
		"staff1@test.com",
		"staff2@test.com",
		"viewer@test.com",
	}

	// Role indices for each user
	roleIndices := []int{0, 1, 2, 2, 3, 3, 4} // Maps to the roles array by index

	// Marina assignments (index in testMarinas array, -1 means no marina as primary)
	marinaIndices := []int{-1, -1, 0, 1, 0, 1, 0}

	// Whether user is superuser
	isSuperuser := []bool{true, false, false, false, false, false, false}

	// Create test password
	testPassword, err := bcrypt.GenerateFromPassword([]byte("testpassword123"), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("failed to encrypt password: %v", err)
	}

	var testUsers []sqlc.User

	for i, username := range usernames {
		user, err := q.GetUserByEmail(ctx, emails[i])
		if err != nil {
			if err == pgx.ErrNoRows {
				// Determine marinaID
				var marinaID uuid.UUID
				if marinaIndices[i] >= 0 && marinaIndices[i] < len(testMarinas) {
					marinaID = testMarinas[marinaIndices[i]].ID
				}

				createParams := sqlc.CreateUserParams{
					Username:       username,
					FirstName:      firstNames[i],
					LastName:       lastNames[i],
					Email:          emails[i],
					PasswordHash:   u.Pointer(string(testPassword)),
					OrganizationID: testOrg.ID,
					RoleID:         testRoles[roleIndices[i]].ID,
					IsSuperuser:    u.Pointer(isSuperuser[i]),
					IsActive:       u.Pointer(true),
				}

				// Only set MarinaID if a marina was assigned
				if marinaIndices[i] >= 0 {
					createParams.MarinaID = marinaID
				}

				user, err = q.CreateUser(ctx, createParams)
				if err != nil {
					log.Fatalf("failed to create test user %s: %v", username, err)
				}
				testUsers = append(testUsers, user)
				log.Printf("Created test user: %s", username)

				// Assign users to marinas in the junction table
				// We'll assign each user to their primary marina and optionally more
				if marinaIndices[i] >= 0 {
					err = q.AssignUserToMarina(ctx, sqlc.AssignUserToMarinaParams{
						UserID:   user.ID,
						MarinaID: marinaID,
					})
					if err != nil {
						log.Fatalf("failed to assign user %s to primary marina: %v", username, err)
					}

					// If user has role that can manage multiple marinas, assign to more than one
					if roleIndices[i] <= 2 { // superuser, org_admin, marina_admin
						// Assign to all marinas
						for _, marina := range testMarinas {
							if marina.ID != marinaID { // Skip primary marina, already assigned
								err = q.AssignUserToMarina(ctx, sqlc.AssignUserToMarinaParams{
									UserID:   user.ID,
									MarinaID: marina.ID,
								})
								if err != nil {
									log.Printf("failed to assign user %s to additional marina: %v", username, err)
								}
							}
						}
					}
				}
			} else {
				log.Fatalf("failed to get user: %v", err)
			}
		} else {
			testUsers = append(testUsers, user)
			log.Printf("Test user already exists: %s", username)
		}
	}

	log.Println("Test seed completed successfully")
}
