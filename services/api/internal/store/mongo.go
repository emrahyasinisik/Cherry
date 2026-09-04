package store

import (
	"context"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const defaultDB = "cherry"

// Mongo implements Store against the Cherry platform database.
type Mongo struct {
	client *mongo.Client
	db     *mongo.Database
}

// OpenMongo connects, pings, ensures indexes, and returns a Store.
// Database name comes from the URI path, or "cherry" when omitted.
func OpenMongo(ctx context.Context, uri string) (*Mongo, error) {
	ctx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}
	if err := client.Ping(ctx, nil); err != nil {
		_ = client.Disconnect(context.Background())
		return nil, err
	}
	name := dbNameFromURI(uri)
	m := &Mongo{client: client, db: client.Database(name)}
	if err := m.ensureIndexes(ctx); err != nil {
		_ = client.Disconnect(context.Background())
		return nil, err
	}
	return m, nil
}

// TryMongo pings MongoDB. Prefer OpenMongo for a full Store.
func TryMongo(ctx context.Context, uri string) (*mongo.Client, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}
	if err := client.Ping(ctx, nil); err != nil {
		_ = client.Disconnect(context.Background())
		return nil, err
	}
	return client, nil
}

func dbNameFromURI(uri string) string {
	// mongodb://host:27017/cherry?opts — take path segment after host.
	rest := uri
	if i := strings.Index(rest, "://"); i >= 0 {
		rest = rest[i+3:]
	}
	if i := strings.Index(rest, "/"); i >= 0 {
		path := rest[i+1:]
		if j := strings.IndexAny(path, "?#"); j >= 0 {
			path = path[:j]
		}
		path = strings.Trim(path, "/")
		if path != "" && !strings.Contains(path, "/") {
			return path
		}
	}
	return defaultDB
}

func (m *Mongo) Name() string { return "mongo" }

func (m *Mongo) DBName() string { return m.db.Name() }

func (m *Mongo) Close(ctx context.Context) error {
	if m == nil || m.client == nil {
		return nil
	}
	return m.client.Disconnect(ctx)
}

func (m *Mongo) Ping(ctx context.Context) error {
	return m.client.Ping(ctx, nil)
}

func (m *Mongo) col(name string) *mongo.Collection {
	return m.db.Collection(name)
}

func (m *Mongo) ensureIndexes(ctx context.Context) error {
	type idx struct {
		coll   string
		models []mongo.IndexModel
	}
	specs := []idx{
		{"users", []mongo.IndexModel{{
			Keys:    bson.D{{Key: "email", Value: 1}},
			Options: options.Index().SetUnique(true),
		}}},
		{"sessions", []mongo.IndexModel{
			{Keys: bson.D{{Key: "tokenHash", Value: 1}}, Options: options.Index().SetUnique(true)},
			{Keys: bson.D{{Key: "userId", Value: 1}}},
		}},
		{"devices", []mongo.IndexModel{{
			Keys:    bson.D{{Key: "userId", Value: 1}, {Key: "fpHash", Value: 1}},
			Options: options.Index().SetUnique(true),
		}}},
		{"verificationCodes", []mongo.IndexModel{
			{Keys: bson.D{{Key: "linkHash", Value: 1}}},
			{Keys: bson.D{{Key: "userId", Value: 1}, {Key: "purpose", Value: 1}}},
			{Keys: bson.D{{Key: "expiresAt", Value: 1}}, Options: options.Index().SetExpireAfterSeconds(0)},
		}},
		{"tempMailboxes", []mongo.IndexModel{
			{Keys: bson.D{{Key: "userId", Value: 1}}},
			{Keys: bson.D{{Key: "challengeId", Value: 1}}},
		}},
		{"projects", []mongo.IndexModel{{
			Keys: bson.D{{Key: "userId", Value: 1}, {Key: "createdAt", Value: -1}},
		}}},
		{"jobs", []mongo.IndexModel{{
			Keys: bson.D{{Key: "projectId", Value: 1}, {Key: "at", Value: 1}},
		}}},
		{"connections", []mongo.IndexModel{{
			Keys:    bson.D{{Key: "userId", Value: 1}, {Key: "kind", Value: 1}},
			Options: options.Index().SetUnique(true),
		}}},
		{"llmVersions", []mongo.IndexModel{{
			Keys: bson.D{{Key: "slot", Value: 1}, {Key: "createdAt", Value: -1}},
		}}},
		{"auditEvents", []mongo.IndexModel{{
			Keys: bson.D{{Key: "userId", Value: 1}, {Key: "createdAt", Value: -1}},
		}}},
	}
	for _, spec := range specs {
		if _, err := m.col(spec.coll).Indexes().CreateMany(ctx, spec.models); err != nil {
			return err
		}
	}
	return nil
}

func (m *Mongo) CreateUser(ctx context.Context, user User) (User, error) {
	email := strings.ToLower(strings.TrimSpace(user.Email))
	if user.ID == "" {
		user.ID = NewID()
	}
	user.Email = email
	doc := bson.M{
		"_id":           user.ID,
		"email":         user.Email,
		"passwordHash":  user.PasswordHash,
		"workspaceKind": string(user.WorkspaceKind),
		"totpSecret":    user.TotpSecret,
		"totpEnabled":   user.TotpEnabled,
	}
	_, err := m.col("users").InsertOne(ctx, doc)
	if mongo.IsDuplicateKeyError(err) {
		return User{}, ErrExists
	}
	if err != nil {
		return User{}, err
	}
	return user, nil
}

func (m *Mongo) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	var doc userDoc
	err := m.col("users").FindOne(ctx, bson.M{"email": strings.ToLower(strings.TrimSpace(email))}).Decode(&doc)
	if err == mongo.ErrNoDocuments {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	u := doc.toUser()
	return &u, nil
}

func (m *Mongo) GetUserByID(ctx context.Context, id string) (*User, error) {
	var doc userDoc
	err := m.col("users").FindOne(ctx, bson.M{"_id": id}).Decode(&doc)
	if err == mongo.ErrNoDocuments {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	u := doc.toUser()
	return &u, nil
}

func (m *Mongo) UpdateUser(ctx context.Context, user User) error {
	res, err := m.col("users").UpdateOne(ctx, bson.M{"_id": user.ID}, bson.M{"$set": bson.M{
		"email":         strings.ToLower(strings.TrimSpace(user.Email)),
		"passwordHash":  user.PasswordHash,
		"workspaceKind": string(user.WorkspaceKind),
		"totpSecret":    user.TotpSecret,
		"totpEnabled":   user.TotpEnabled,
	}})
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return ErrNotFound
	}
	return nil
}

func (m *Mongo) PutChallenge(ctx context.Context, challenge Challenge) error {
	doc := bson.M{
		"_id":          challenge.ID,
		"userId":       challenge.UserID,
		"purpose":      string(challenge.Purpose),
		"codeHash":     challenge.CodeHash,
		"linkHash":     challenge.LinkHash,
		"attempts":     challenge.Attempts,
		"maxAttempts":  challenge.MaxAttempts,
		"expiresAt":    challenge.ExpiresAt,
		"deviceFpHash": challenge.DeviceFPHash,
		"deviceLabel":  challenge.DeviceLabel,
		"codeVerified": challenge.CodeVerified,
		"consumed":     challenge.Consumed,
		"trustDevice":  challenge.TrustDevice,
	}
	_, err := m.col("verificationCodes").ReplaceOne(ctx, bson.M{"_id": challenge.ID}, doc, options.Replace().SetUpsert(true))
	return err
}

func (m *Mongo) GetChallenge(ctx context.Context, id string) (*Challenge, error) {
	var doc challengeDoc
	err := m.col("verificationCodes").FindOne(ctx, bson.M{"_id": id}).Decode(&doc)
	if err == mongo.ErrNoDocuments {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	c := doc.toChallenge()
	return &c, nil
}

func (m *Mongo) GetChallengeByLinkHash(ctx context.Context, linkHash string) (*Challenge, error) {
	var doc challengeDoc
	err := m.col("verificationCodes").FindOne(ctx, bson.M{"linkHash": linkHash, "consumed": false}).Decode(&doc)
	if err == mongo.ErrNoDocuments {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	c := doc.toChallenge()
	return &c, nil
}

func (m *Mongo) InvalidateChallenges(ctx context.Context, userID string, purpose Purpose) error {
	_, err := m.col("verificationCodes").UpdateMany(ctx,
		bson.M{"userId": userID, "purpose": string(purpose), "consumed": false},
		bson.M{"$set": bson.M{"consumed": true}},
	)
	return err
}

func (m *Mongo) CreateSession(ctx context.Context, session Session) error {
	doc := bson.M{
		"_id":         session.ID,
		"userId":      session.UserID,
		"tokenHash":   session.TokenHash,
		"deviceId":    session.DeviceID,
		"deviceLabel": session.DeviceLabel,
		"createdAt":   session.CreatedAt,
		"revoked":     session.Revoked,
	}
	_, err := m.col("sessions").InsertOne(ctx, doc)
	return err
}

func (m *Mongo) GetSessionByTokenHash(ctx context.Context, tokenHash string) (*Session, error) {
	var doc sessionDoc
	err := m.col("sessions").FindOne(ctx, bson.M{"tokenHash": tokenHash, "revoked": false}).Decode(&doc)
	if err == mongo.ErrNoDocuments {
		return nil, ErrUnauthorized
	}
	if err != nil {
		return nil, err
	}
	s := doc.toSession()
	return &s, nil
}

func (m *Mongo) ListSessions(ctx context.Context, userID string) ([]Session, error) {
	cur, err := m.col("sessions").Find(ctx, bson.M{"userId": userID, "revoked": false})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	out := make([]Session, 0)
	for cur.Next(ctx) {
		var doc sessionDoc
		if err := cur.Decode(&doc); err != nil {
			return nil, err
		}
		out = append(out, doc.toSession())
	}
	return out, cur.Err()
}

func (m *Mongo) RevokeSession(ctx context.Context, id string) error {
	res, err := m.col("sessions").UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": bson.M{"revoked": true}})
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return ErrNotFound
	}
	return nil
}

func (m *Mongo) RevokeOtherSessions(ctx context.Context, userID, keepID string) error {
	_, err := m.col("sessions").UpdateMany(ctx,
		bson.M{"userId": userID, "_id": bson.M{"$ne": keepID}},
		bson.M{"$set": bson.M{"revoked": true}},
	)
	return err
}

func (m *Mongo) UpsertDevice(ctx context.Context, device Device) (Device, error) {
	filter := bson.M{"userId": device.UserID, "fpHash": device.FPHash}
	var existing deviceDoc
	err := m.col("devices").FindOne(ctx, filter).Decode(&existing)
	if err == nil {
		existing.Label = device.Label
		existing.Trusted = device.Trusted
		existing.LastSeen = device.LastSeen
		_, err = m.col("devices").ReplaceOne(ctx, bson.M{"_id": existing.ID}, existing)
		if err != nil {
			return Device{}, err
		}
		return existing.toDevice(), nil
	}
	if err != mongo.ErrNoDocuments {
		return Device{}, err
	}
	if device.ID == "" {
		device.ID = NewID()
	}
	doc := deviceDoc{
		ID:       device.ID,
		UserID:   device.UserID,
		FPHash:   device.FPHash,
		Label:    device.Label,
		Trusted:  device.Trusted,
		LastSeen: device.LastSeen,
	}
	_, err = m.col("devices").InsertOne(ctx, doc)
	if err != nil {
		return Device{}, err
	}
	return device, nil
}

func (m *Mongo) GetDeviceByFP(ctx context.Context, userID, fpHash string) (*Device, error) {
	var doc deviceDoc
	err := m.col("devices").FindOne(ctx, bson.M{"userId": userID, "fpHash": fpHash}).Decode(&doc)
	if err == mongo.ErrNoDocuments {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	d := doc.toDevice()
	return &d, nil
}

func (m *Mongo) ListDevices(ctx context.Context, userID string) ([]Device, error) {
	cur, err := m.col("devices").Find(ctx, bson.M{"userId": userID})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	out := make([]Device, 0)
	for cur.Next(ctx) {
		var doc deviceDoc
		if err := cur.Decode(&doc); err != nil {
			return nil, err
		}
		out = append(out, doc.toDevice())
	}
	return out, cur.Err()
}

func (m *Mongo) RevokeDevice(ctx context.Context, id string) error {
	res, err := m.col("devices").UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": bson.M{"trusted": false}})
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return ErrNotFound
	}
	return nil
}

func (m *Mongo) AddMail(ctx context.Context, mail Mail) error {
	if mail.ID == "" {
		mail.ID = NewID()
	}
	doc := bson.M{
		"_id":         mail.ID,
		"userId":      mail.UserID,
		"challengeId": mail.ChallengeID,
		"subject":     mail.Subject,
		"body":        mail.Body,
		"purpose":     string(mail.Purpose),
		"createdAt":   mail.CreatedAt,
	}
	_, err := m.col("tempMailboxes").InsertOne(ctx, doc)
	return err
}

func (m *Mongo) ListMail(ctx context.Context, userID string) ([]Mail, error) {
	cur, err := m.col("tempMailboxes").Find(ctx, bson.M{"userId": userID})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	out := make([]Mail, 0)
	for cur.Next(ctx) {
		var doc mailDoc
		if err := cur.Decode(&doc); err != nil {
			return nil, err
		}
		out = append(out, doc.toMail())
	}
	return out, cur.Err()
}

func (m *Mongo) MailByChallenge(ctx context.Context, challengeID string) (*Mail, error) {
	opts := options.FindOne().SetSort(bson.D{{Key: "createdAt", Value: -1}})
	var doc mailDoc
	err := m.col("tempMailboxes").FindOne(ctx, bson.M{"challengeId": challengeID}, opts).Decode(&doc)
	if err == mongo.ErrNoDocuments {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	mail := doc.toMail()
	return &mail, nil
}

func (m *Mongo) Projects(ctx context.Context, userID string) ([]Project, error) {
	opts := options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}})
	cur, err := m.col("projects").Find(ctx, bson.M{"userId": userID}, opts)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	out := make([]Project, 0)
	for cur.Next(ctx) {
		var doc projectDoc
		if err := cur.Decode(&doc); err != nil {
			return nil, err
		}
		out = append(out, doc.toProject())
	}
	return out, cur.Err()
}

func (m *Mongo) CreateProject(ctx context.Context, project Project) (Project, error) {
	if project.ID == "" {
		project.ID = NewID()
	}
	doc := projectDoc{
		ID:        project.ID,
		UserID:    project.UserID,
		Name:      project.Name,
		Brief:     project.Brief,
		Stack:     string(project.Stack),
		Status:    string(project.Status),
		RootPath:  project.RootPath,
		Backend:   string(project.Backend),
		CreatedAt: project.CreatedAt,
		UpdatedAt: project.UpdatedAt,
	}
	_, err := m.col("projects").InsertOne(ctx, doc)
	if err != nil {
		return Project{}, err
	}
	return project, nil
}

func (m *Mongo) GetProject(ctx context.Context, id string) (*Project, error) {
	var doc projectDoc
	err := m.col("projects").FindOne(ctx, bson.M{"_id": id}).Decode(&doc)
	if err == mongo.ErrNoDocuments {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	p := doc.toProject()
	return &p, nil
}

func (m *Mongo) UpdateProject(ctx context.Context, project Project) error {
	res, err := m.col("projects").ReplaceOne(ctx, bson.M{"_id": project.ID}, projectDoc{
		ID:        project.ID,
		UserID:    project.UserID,
		Name:      project.Name,
		Brief:     project.Brief,
		Stack:     string(project.Stack),
		Status:    string(project.Status),
		RootPath:  project.RootPath,
		Backend:   string(project.Backend),
		CreatedAt: project.CreatedAt,
		UpdatedAt: project.UpdatedAt,
	})
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return ErrNotFound
	}
	return nil
}

func (m *Mongo) AppendLog(ctx context.Context, log JobLog) error {
	if log.ID == "" {
		log.ID = NewID()
	}
	_, err := m.col("jobs").InsertOne(ctx, bson.M{
		"_id":       log.ID,
		"projectId": log.ProjectID,
		"at":        log.At,
		"message":   log.Message,
		"role":      string(log.Role),
	})
	return err
}

func (m *Mongo) ListLogs(ctx context.Context, projectID string) ([]JobLog, error) {
	opts := options.Find().SetSort(bson.D{{Key: "at", Value: 1}})
	cur, err := m.col("jobs").Find(ctx, bson.M{"projectId": projectID}, opts)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	out := make([]JobLog, 0)
	for cur.Next(ctx) {
		var doc jobLogDoc
		if err := cur.Decode(&doc); err != nil {
			return nil, err
		}
		out = append(out, doc.toJobLog())
	}
	return out, cur.Err()
}

func (m *Mongo) GetConnection(ctx context.Context, userID string, kind ConnectionKind) (*Connection, error) {
	var doc connectionDoc
	err := m.col("connections").FindOne(ctx, bson.M{"userId": userID, "kind": string(kind)}).Decode(&doc)
	if err == mongo.ErrNoDocuments {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	c := doc.toConnection()
	return &c, nil
}

func (m *Mongo) UpsertConnection(ctx context.Context, conn Connection) (Connection, error) {
	filter := bson.M{"userId": conn.UserID, "kind": string(conn.Kind)}
	var existing connectionDoc
	err := m.col("connections").FindOne(ctx, filter).Decode(&existing)
	switch {
	case err == nil:
		conn.ID = existing.ID
	case err == mongo.ErrNoDocuments:
		if conn.ID == "" {
			conn.ID = NewID()
		}
	default:
		return Connection{}, err
	}
	doc := connectionDoc{
		ID:         conn.ID,
		UserID:     conn.UserID,
		Kind:       string(conn.Kind),
		Status:     string(conn.Status),
		Account:    conn.Account,
		Token:      conn.Token,
		TokenHint:  conn.TokenHint,
		Note:       conn.Note,
		AuthMethod: string(conn.AuthMethod),
		Scopes:     conn.Scopes,
		UpdatedAt:  conn.UpdatedAt,
	}
	_, err = m.col("connections").ReplaceOne(ctx, filter, doc, options.Replace().SetUpsert(true))
	if err != nil {
		return Connection{}, err
	}
	return conn, nil
}

func (m *Mongo) DeleteConnection(ctx context.Context, userID string, kind ConnectionKind) error {
	_, err := m.col("connections").DeleteOne(ctx, bson.M{"userId": userID, "kind": string(kind)})
	return err
}

func (m *Mongo) ListConnections(ctx context.Context, userID string) ([]Connection, error) {
	cur, err := m.col("connections").Find(ctx, bson.M{"userId": userID})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	out := make([]Connection, 0)
	for cur.Next(ctx) {
		var doc connectionDoc
		if err := cur.Decode(&doc); err != nil {
			return nil, err
		}
		out = append(out, doc.toConnection())
	}
	return out, cur.Err()
}

func (m *Mongo) PutLlmVersion(ctx context.Context, version LlmVersion) error {
	if version.ID == "" {
		version.ID = NewID()
	}
	_, err := m.col("llmVersions").ReplaceOne(ctx, bson.M{"_id": version.ID}, bson.M{
		"_id":           version.ID,
		"slot":          string(version.Slot),
		"name":          version.Name,
		"note":          version.Note,
		"checkpointRef": version.CheckpointRef,
		"createdAt":     version.CreatedAt,
	}, options.Replace().SetUpsert(true))
	return err
}

func (m *Mongo) ListLlmVersions(ctx context.Context, slot LlmSlot) ([]LlmVersion, error) {
	cur, err := m.col("llmVersions").Find(ctx, bson.M{"slot": string(slot)})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	out := make([]LlmVersion, 0)
	for cur.Next(ctx) {
		var doc llmVersionDoc
		if err := cur.Decode(&doc); err != nil {
			return nil, err
		}
		out = append(out, doc.toLlmVersion())
	}
	return out, cur.Err()
}

func (m *Mongo) GetLlmVersion(ctx context.Context, id string) (*LlmVersion, error) {
	var doc llmVersionDoc
	err := m.col("llmVersions").FindOne(ctx, bson.M{"_id": id}).Decode(&doc)
	if err == mongo.ErrNoDocuments {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	v := doc.toLlmVersion()
	return &v, nil
}

func (m *Mongo) GetLlmState(ctx context.Context) (LlmState, error) {
	var doc llmStateDoc
	err := m.col("llmState").FindOne(ctx, bson.M{"_id": "singleton"}).Decode(&doc)
	if err == mongo.ErrNoDocuments {
		return LlmState{}, nil
	}
	if err != nil {
		return LlmState{}, err
	}
	return LlmState{
		ActiveAID: doc.ActiveAID,
		ActiveBID: doc.ActiveBID,
		McpRoot:   doc.McpRoot,
		ColabURLA: doc.ColabURLA,
		ColabURLB: doc.ColabURLB,
	}, nil
}

func (m *Mongo) SetLlmState(ctx context.Context, state LlmState) error {
	_, err := m.col("llmState").ReplaceOne(ctx, bson.M{"_id": "singleton"}, bson.M{
		"_id":       "singleton",
		"activeAId": state.ActiveAID,
		"activeBId": state.ActiveBID,
		"mcpRoot":   state.McpRoot,
		"colabUrlA": state.ColabURLA,
		"colabUrlB": state.ColabURLB,
	}, options.Replace().SetUpsert(true))
	return err
}

func (m *Mongo) AddAudit(ctx context.Context, event AuditEvent) error {
	if event.ID == "" {
		event.ID = NewID()
	}
	_, err := m.col("auditEvents").InsertOne(ctx, bson.M{
		"_id":              event.ID,
		"userId":           event.UserID,
		"projectId":        event.ProjectID,
		"purpose":          event.Purpose,
		"legalBasis":       event.LegalBasis,
		"slot":             string(event.Slot),
		"versionId":        event.VersionID,
		"versionName":      event.VersionName,
		"channel":          event.Channel,
		"inputRedactions":  event.InputRedactions,
		"outputRedactions": event.OutputRedactions,
		"promptPreview":    event.PromptPreview,
		"outputPreview":    event.OutputPreview,
		"createdAt":        event.CreatedAt,
	})
	return err
}

func (m *Mongo) ListAudit(ctx context.Context, userID string) ([]AuditEvent, error) {
	opts := options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}})
	cur, err := m.col("auditEvents").Find(ctx, bson.M{"userId": userID}, opts)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	out := make([]AuditEvent, 0)
	for cur.Next(ctx) {
		var doc auditDoc
		if err := cur.Decode(&doc); err != nil {
			return nil, err
		}
		out = append(out, doc.toAudit())
	}
	return out, cur.Err()
}

func (m *Mongo) DeleteUserData(ctx context.Context, userID string, wipeProjects bool) error {
	res, err := m.col("users").DeleteOne(ctx, bson.M{"_id": userID})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return ErrNotFound
	}
	_, _ = m.col("sessions").DeleteMany(ctx, bson.M{"userId": userID})
	_, _ = m.col("devices").DeleteMany(ctx, bson.M{"userId": userID})
	_, _ = m.col("tempMailboxes").DeleteMany(ctx, bson.M{"userId": userID})
	_, _ = m.col("connections").DeleteMany(ctx, bson.M{"userId": userID})
	_, _ = m.col("verificationCodes").DeleteMany(ctx, bson.M{"userId": userID})
	_, _ = m.col("auditEvents").DeleteMany(ctx, bson.M{"userId": userID})
	if wipeProjects {
		cur, err := m.col("projects").Find(ctx, bson.M{"userId": userID}, options.Find().SetProjection(bson.M{"_id": 1}))
		if err != nil {
			return err
		}
		ids := make([]string, 0)
		for cur.Next(ctx) {
			var row struct {
				ID string `bson:"_id"`
			}
			if err := cur.Decode(&row); err != nil {
				_ = cur.Close(ctx)
				return err
			}
			ids = append(ids, row.ID)
		}
		_ = cur.Close(ctx)
		if len(ids) > 0 {
			_, _ = m.col("jobs").DeleteMany(ctx, bson.M{"projectId": bson.M{"$in": ids}})
		}
		_, _ = m.col("projects").DeleteMany(ctx, bson.M{"userId": userID})
	}
	return nil
}

// Compile-time check.
var _ Store = (*Mongo)(nil)
