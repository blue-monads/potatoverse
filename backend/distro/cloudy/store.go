package cloudy

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/upper/db/v4"
	"github.com/upper/db/v4/adapter/sqlite"
)

type Store struct {
	sess db.Session
	sql  *sql.DB
}

func newStore(sqlDB *sql.DB) (*Store, error) {
	sess, err := sqlite.New(sqlDB)
	if err != nil {
		return nil, err
	}
	return &Store{sess: sess, sql: sqlDB}, nil
}

func (s *Store) users() db.Collection {
	return s.sess.Collection("CloudyUsers")
}

func (s *Store) teams() db.Collection {
	return s.sess.Collection("CloudyTeams")
}

func (s *Store) teamMembers() db.Collection {
	return s.sess.Collection("CloudyTeamMembers")
}

func (s *Store) potatoInstances() db.Collection {
	return s.sess.Collection("CloudyTeamPotatoInstances")
}

func (s *Store) listUsers() ([]*User, error) {
	var users []*User
	err := s.users().Find().OrderBy("id").All(&users)
	if err != nil {
		return nil, err
	}
	for _, u := range users {
		normalizeUser(u)
	}
	return users, nil
}

func (s *Store) getUserByID(id int64) (*User, error) {
	u := &User{}
	err := s.users().Find(db.Cond{"id": id}).One(u)
	if errors.Is(err, db.ErrNoMoreRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	normalizeUser(u)
	return u, nil
}

func (s *Store) getUserByEmail(email string) (*User, error) {
	u := &User{}
	err := s.users().Find(db.Cond{"email": email}).One(u)
	if errors.Is(err, db.ErrNoMoreRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	normalizeUser(u)
	return u, nil
}

func (s *Store) insertUserWithTeam(fullname, email, passwordHash, teamName, utype string, verified bool) (*User, *Team, error) {
	if utype == "" {
		utype = UTypeNormal
	}
	if !validUType(utype) {
		return nil, nil, fmt.Errorf("utype must be admin or normal")
	}

	// Check if email already exists
	existingUser, err := s.getUserByEmail(email)
	if err != nil {
		return nil, nil, err
	}
	if existingUser != nil {
		return nil, nil, fmt.Errorf("email already in use")
	}

	now := time.Now().UTC()
	u := &User{
		Fullname:   fullname,
		Email:      email,
		Password:   passwordHash,
		UType:      utype,
		IsVerified: verified,
		IsDisabled: false,
		CreatedAt:  &now,
		UpdatedAt:  &now,
	}

	_, err = s.users().Insert(u)
	if err != nil {
		return nil, nil, fmt.Errorf("insert user: %w", err)
	}

	createdUser, err := s.getUserByEmail(email)
	if err != nil || createdUser == nil {
		return nil, nil, fmt.Errorf("read back created user for %s: %w", email, err)
	}
	userID := createdUser.ID

	if teamName == "" {
		teamName = fmt.Sprintf("%s's Team", fullname)
	}

	team := &Team{
		Name:        teamName,
		Description: fmt.Sprintf("%s's Team", fullname),
		OwnerID:     userID,
		IsDeleted:   false,
		CreatedAt:   &now,
		UpdatedAt:   &now,
	}
	_, err = s.teams().Insert(team)
	if err != nil {
		return nil, nil, fmt.Errorf("insert team: %w", err)
	}

	var createdTeam Team
	err = s.teams().Find(db.Cond{"name": teamName, "owner_id": userID, "is_deleted": false}).OrderBy("id DESC").One(&createdTeam)
	if err != nil || createdTeam.ID == 0 {
		return nil, nil, fmt.Errorf("read back created team for user %d: %w", userID, err)
	}
	teamID := createdTeam.ID

	member := &TeamMember{
		TeamID:    teamID,
		UserID:    userID,
		IsDeleted: false,
		CreatedAt: &now,
		UpdatedAt: &now,
	}
	if _, err := s.teamMembers().Insert(member); err != nil {
		return nil, nil, fmt.Errorf("insert team member: %w", err)
	}

	return createdUser, &createdTeam, nil
}

func (s *Store) createPotatoInstance(teamID int64, slug, description string) (*TeamPotatoInstance, error) {
	inst, err := s.getPotatoInstanceBySlug(slug)
	if err != nil {
		return nil, err
	}
	if inst != nil {
		return nil, fmt.Errorf("instance slug already in use")
	}

	now := time.Now().UTC()
	if description == "" {
		description = fmt.Sprintf("%s Potato Instance", slug)
	}

	instance := &TeamPotatoInstance{
		Slug:        slug,
		Description: description,
		TeamID:      teamID,
		IsDeleted:   false,
		CreatedAt:   &now,
		UpdatedAt:   &now,
	}
	_, err = s.potatoInstances().Insert(instance)
	if err != nil {
		return nil, fmt.Errorf("insert potato instance: %w", err)
	}

	createdInst, err := s.getPotatoInstanceBySlug(slug)
	if err != nil || createdInst == nil {
		return nil, fmt.Errorf("read back created instance for %s: %w", slug, err)
	}
	return createdInst, nil
}

func (s *Store) markUserVerified(id int64) error {
	now := time.Now().UTC()
	return s.users().Find(db.Cond{"id": id}).Update(map[string]any{
		"is_verified": true,
		"updated_at":  now,
	})
}

func (s *Store) updateUserPassword(id int64, passwordHash string) error {
	now := time.Now().UTC()
	return s.users().Find(db.Cond{"id": id}).Update(map[string]any{
		"password":   passwordHash,
		"updated_at": now,
	})
}

func (s *Store) setUserDisabled(id int64, disabled bool) error {
	now := time.Now().UTC()
	return s.users().Find(db.Cond{"id": id}).Update(map[string]any{
		"is_disabled": disabled,
		"updated_at":  now,
	})
}

func (s *Store) listTeams() ([]*Team, error) {
	var teams []*Team
	err := s.teams().Find(db.Cond{"is_deleted": false}).OrderBy("id").All(&teams)
	if err != nil {
		return nil, err
	}
	return teams, nil
}

func (s *Store) getTeamByID(id int64) (*Team, error) {
	t := &Team{}
	err := s.teams().Find(db.Cond{"id": id, "is_deleted": false}).One(t)
	if errors.Is(err, db.ErrNoMoreRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (s *Store) getTeamsForUser(userID int64) ([]*Team, error) {
	var members []*TeamMember
	err := s.teamMembers().Find(db.Cond{"user_id": userID, "is_deleted": false}).All(&members)
	if err != nil {
		return nil, err
	}
	teams := make([]*Team, 0, len(members))
	for _, m := range members {
		t, err := s.getTeamByID(m.TeamID)
		if err == nil && t != nil {
			teams = append(teams, t)
		}
	}
	return teams, nil
}

func (s *Store) isUserInTeam(userID, teamID int64) (bool, error) {
	n, err := s.teamMembers().Find(db.Cond{"user_id": userID, "team_id": teamID, "is_deleted": false}).Count()
	return n > 0, err
}

func (s *Store) getPotatoInstanceBySlug(slug string) (*TeamPotatoInstance, error) {
	inst := &TeamPotatoInstance{}
	err := s.potatoInstances().Find(db.Cond{"slug": slug, "is_deleted": false}).One(inst)
	if errors.Is(err, db.ErrNoMoreRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return inst, nil
}

func (s *Store) getPotatoInstancesByTeamID(teamID int64) ([]*TeamPotatoInstance, error) {
	var insts []*TeamPotatoInstance
	err := s.potatoInstances().Find(db.Cond{"team_id": teamID, "is_deleted": false}).OrderBy("id").All(&insts)
	if err != nil {
		return nil, err
	}
	return insts, nil
}

func (s *Store) listPotatoInstances() ([]*TeamPotatoInstance, error) {
	var insts []*TeamPotatoInstance
	err := s.potatoInstances().Find(db.Cond{"is_deleted": false}).OrderBy("id").All(&insts)
	if err != nil {
		return nil, err
	}
	return insts, nil
}

func (s *Store) getPrimaryInstanceForUser(userID int64) (*TeamPotatoInstance, *Team, error) {
	var members []*TeamMember
	err := s.teamMembers().Find(db.Cond{"user_id": userID, "is_deleted": false}).All(&members)
	if err != nil {
		return nil, nil, err
	}
	if len(members) == 0 {
		return nil, nil, nil
	}

	for _, m := range members {
		team, err := s.getTeamByID(m.TeamID)
		if err != nil || team == nil {
			continue
		}
		var insts []*TeamPotatoInstance
		err = s.potatoInstances().Find(db.Cond{"team_id": team.ID, "is_deleted": false}).All(&insts)
		if err == nil && len(insts) > 0 {
			return insts[0], team, nil
		}
	}
	return nil, nil, nil
}

func (s *Store) slugExists(slug string) bool {
	inst, err := s.getPotatoInstanceBySlug(slug)
	return err == nil && inst != nil
}

func normalizeUser(u *User) {
	if u == nil {
		return
	}
	if u.UType == "" {
		u.UType = UTypeNormal
	}
}
