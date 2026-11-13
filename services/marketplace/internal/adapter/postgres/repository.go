package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/your-org/app/marketplace/internal/domain"
	"golang.org/x/crypto/bcrypt"
	"time"
)

type Repository struct{ DB *pgxpool.Pool }

func translate(e error) error {
	if errors.Is(e, pgx.ErrNoRows) {
		return domain.ErrNotFound
	}
	return e
}
func scanUser(row pgx.Row) (u domain.User, e error) {
	e = row.Scan(&u.ID, &u.Email, &u.Role, &u.Name, &u.Phone, &u.Provider)
	return u, translate(e)
}

const userColumns = "id::text,email,role,name,phone,provider"

func (r *Repository) Allow(ctx context.Context, key string, limit int, duration time.Duration) (bool, error) {
	var n int
	e := r.DB.QueryRow(ctx, `INSERT INTO market_limits(key,count,until_at) VALUES($1,1,now()+$2*interval '1 second') ON CONFLICT(key) DO UPDATE SET count=CASE WHEN market_limits.until_at<now() THEN 1 ELSE market_limits.count+1 END,until_at=CASE WHEN market_limits.until_at<now() THEN excluded.until_at ELSE market_limits.until_at END RETURNING count`, key, int(duration.Seconds())).Scan(&n)
	return n <= limit, e
}
func (r *Repository) CreateChallenge(ctx context.Context, id, email, role, hash string) error {
	tag, e := r.DB.Exec(ctx, `INSERT INTO market_challenges(id,email,role,code_hash,expires_at) VALUES($1,$2,$3,$4,now()+interval '10 minutes') ON CONFLICT(email) DO UPDATE SET id=excluded.id,role=excluded.role,code_hash=excluded.code_hash,attempts=0,expires_at=excluded.expires_at,created_at=now() WHERE market_challenges.created_at<now()-interval '60 seconds'`, id, email, role, hash)
	if e == nil && tag.RowsAffected() == 0 {
		return domain.ErrRateLimit
	}
	return e
}
func (r *Repository) RemoveChallenge(ctx context.Context, id string) error {
	_, e := r.DB.Exec(ctx, "DELETE FROM market_challenges WHERE id=$1", id)
	return e
}
func (r *Repository) ConsumeChallenge(ctx context.Context, id, hash string) (string, string, error) {
	tx, e := r.DB.Begin(ctx)
	if e != nil {
		return "", "", e
	}
	defer tx.Rollback(ctx)
	var email, role, stored string
	var attempts int
	var valid bool
	e = tx.QueryRow(ctx, `SELECT email,role,code_hash,attempts,expires_at>now() FROM market_challenges WHERE id=$1 FOR UPDATE`, id).Scan(&email, &role, &stored, &attempts, &valid)
	if errors.Is(e, pgx.ErrNoRows) {
		return "", "", domain.ErrUnauthorized
	}
	if e != nil {
		return "", "", e
	}
	if !valid || attempts >= 5 {
		return "", "", domain.ErrUnauthorized
	}
	if stored != hash {
		if _, e = tx.Exec(ctx, "UPDATE market_challenges SET attempts=attempts+1 WHERE id=$1", id); e != nil {
			return "", "", e
		}
		if e = tx.Commit(ctx); e != nil {
			return "", "", e
		}
		return "", "", domain.ErrUnauthorized
	}
	if _, e = tx.Exec(ctx, "DELETE FROM market_challenges WHERE id=$1", id); e != nil {
		return "", "", e
	}
	e = tx.Commit(ctx)
	return email, role, e
}
func (r *Repository) Login(ctx context.Context, email, role, provider, hash string) (domain.User, error) {
	tx, e := r.DB.Begin(ctx)
	if e != nil {
		return domain.User{}, e
	}
	defer tx.Rollback(ctx)
	_, e = tx.Exec(ctx, `INSERT INTO market_users(email,role,provider) VALUES($1,$2,$3) ON CONFLICT(email) DO NOTHING`, email, role, provider)
	if e != nil {
		return domain.User{}, e
	}
	u, e := scanUser(tx.QueryRow(ctx, "SELECT "+userColumns+" FROM market_users WHERE email=$1 AND NOT suspended FOR UPDATE", email))
	if e != nil {
		return u, e
	}
	var passwordAccount bool
	if e = tx.QueryRow(ctx, "SELECT password_hash IS NOT NULL FROM market_users WHERE id=$1", u.ID).Scan(&passwordAccount); e != nil {
		return u, e
	}
	if passwordAccount {
		return domain.User{}, domain.ErrUnauthorized
	}
	if u.Role != role {
		return u, domain.ErrConflict
	}
	if _, e = tx.Exec(ctx, `INSERT INTO market_sessions(token_hash,user_id,expires_at) VALUES($1,$2,now()+interval '30 days')`, hash, u.ID); e != nil {
		return u, e
	}
	_, e = tx.Exec(ctx, "UPDATE market_users SET provider=$1 WHERE id=$2", provider, u.ID)
	if e != nil {
		return u, e
	}
	u.Provider = provider
	return u, tx.Commit(ctx)
}
func (r *Repository) Authenticate(ctx context.Context, hash string) (domain.User, error) {
	u, e := scanUser(r.DB.QueryRow(ctx, `SELECT u.id::text,u.email,u.role,u.name,u.phone,u.provider FROM market_users u JOIN market_sessions s ON s.user_id=u.id WHERE s.token_hash=$1 AND s.expires_at>now() AND NOT u.suspended`, hash))
	if errors.Is(e, domain.ErrNotFound) {
		e = domain.ErrUnauthorized
	}
	return u, e
}
func (r *Repository) Logout(ctx context.Context, hash string) error {
	_, e := r.DB.Exec(ctx, "DELETE FROM market_sessions WHERE token_hash=$1", hash)
	return e
}
func (r *Repository) SaveProfile(ctx context.Context, id, name, phone string) (domain.User, error) {
	return scanUser(r.DB.QueryRow(ctx, "UPDATE market_users SET name=$1,phone=$2 WHERE id=$3 RETURNING "+userColumns, name, phone, id))
}
func (r *Repository) DeleteAccount(ctx context.Context, id, hash string) error {
	tag, e := r.DB.Exec(ctx, `DELETE FROM market_users WHERE id=$1 AND EXISTS(SELECT 1 FROM market_sessions WHERE token_hash=$2 AND user_id=$1 AND created_at>now()-interval '15 minutes')`, id, hash)
	if e == nil && tag.RowsAffected() == 0 {
		return domain.ErrUnauthorized
	}
	return e
}
func scanProduct(row pgx.Row) (p domain.Product, e error) {
	e = row.Scan(&p.ID, &p.Name, &p.Category, &p.Price, &p.Stock, &p.Description, &p.SellerID, &p.Seller)
	return p, translate(e)
}

const productQuery = `SELECT p.id,p.name,p.category,p.price,p.stock,p.description,p.seller_id::text,COALESCE(NULLIF(u.name,''),'Касапчы') FROM market_products p JOIN market_users u ON u.id=p.seller_id`

func (r *Repository) Products(ctx context.Context) ([]domain.Product, error) {
	rows, e := r.DB.Query(ctx, productQuery+" WHERE NOT p.hidden AND NOT u.suspended ORDER BY p.id DESC LIMIT 500")
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	list := []domain.Product{}
	for rows.Next() {
		p, e := scanProduct(rows)
		if e != nil {
			return nil, e
		}
		list = append(list, p)
	}
	return list, rows.Err()
}
func (r *Repository) Publish(ctx context.Context, u domain.User, p domain.NewProduct) (domain.Product, error) {
	if !p.AcceptedContentPolicy {
		return domain.Product{}, domain.ErrForbidden
	}
	if _, err := r.DB.Exec(ctx, "UPDATE market_users SET content_policy_at=now() WHERE id=$1", u.ID); err != nil {
		return domain.Product{}, err
	}
	var id int64
	e := r.DB.QueryRow(ctx, `INSERT INTO market_products(seller_id,name,category,price,stock,description) VALUES($1,$2,$3,$4,$5,$6) RETURNING id`, u.ID, p.Name, p.Category, p.Price, p.Stock, p.Description).Scan(&id)
	if e != nil {
		return domain.Product{}, e
	}
	return scanProduct(r.DB.QueryRow(ctx, productQuery+" WHERE p.id=$1", id))
}
func scanOrder(row pgx.Row) (o domain.Order, e error) {
	var snapshot []byte
	e = row.Scan(&o.ID, &snapshot, &o.Quantity, &o.BuyerEmail, &o.Delivery, &o.Address, &o.Status, &o.Total, &o.CreatedAt)
	if e != nil {
		return o, translate(e)
	}
	e = json.Unmarshal(snapshot, &o.Product)
	return o, e
}

const orderQuery = `SELECT o.id,o.product_snapshot,o.quantity,u.email,o.delivery,o.address,o.status,o.total,o.created_at FROM market_orders o JOIN market_users u ON u.id=o.buyer_id`

func (r *Repository) Orders(ctx context.Context, u domain.User) ([]domain.Order, error) {
	column := "o.buyer_id"
	if u.Role == "Seller" {
		column = "o.seller_id"
	}
	rows, e := r.DB.Query(ctx, orderQuery+" WHERE "+column+"=$1 ORDER BY o.id DESC LIMIT 500", u.ID)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	list := []domain.Order{}
	for rows.Next() {
		o, e := scanOrder(rows)
		if e != nil {
			return nil, e
		}
		list = append(list, o)
	}
	return list, rows.Err()
}
func (r *Repository) Place(ctx context.Context, u domain.User, input domain.NewOrder) (domain.Order, error) {
	tx, e := r.DB.Begin(ctx)
	if e != nil {
		return domain.Order{}, e
	}
	defer tx.Rollback(ctx)
	// Serialise idempotency keys within a buyer, including concurrent retries.
	if _, e = tx.Exec(ctx, "SELECT id FROM market_users WHERE id=$1 FOR UPDATE", u.ID); e != nil {
		return domain.Order{}, e
	}
	existing, e := scanOrder(tx.QueryRow(ctx, orderQuery+" WHERE o.buyer_id=$1 AND o.request_id=$2", u.ID, input.RequestID))
	if e == nil {
		if existing.Product.ID != input.ProductID || existing.Quantity != input.Quantity || existing.Delivery != input.Delivery || existing.Address != input.Address {
			return domain.Order{}, domain.ErrConflict
		}
		return existing, tx.Commit(ctx)
	}
	if !errors.Is(e, domain.ErrNotFound) {
		return domain.Order{}, e
	}
	p, e := scanProduct(tx.QueryRow(ctx, productQuery+" WHERE p.id=$1 AND NOT p.hidden AND NOT u.suspended AND NOT EXISTS(SELECT 1 FROM market_blocks b WHERE b.user_id=$2 AND b.seller_id=p.seller_id) FOR UPDATE OF p", input.ProductID, u.ID))
	if e != nil {
		return domain.Order{}, e
	}
	if p.Stock < input.Quantity {
		return domain.Order{}, domain.ErrConflict
	}
	total := p.Price * input.Quantity
	if input.Delivery == "Courier" {
		total += 150
	}
	snapshot, e := json.Marshal(p)
	if e != nil {
		return domain.Order{}, e
	}
	var id int64
	e = tx.QueryRow(ctx, `INSERT INTO market_orders(buyer_id,seller_id,product_snapshot,quantity,delivery,address,total,request_id) VALUES($1,$2,$3,$4,$5,$6,$7,$8) RETURNING id`, u.ID, p.SellerID, snapshot, input.Quantity, input.Delivery, input.Address, total, input.RequestID).Scan(&id)
	if e != nil {
		return domain.Order{}, e
	}
	if _, e = tx.Exec(ctx, "UPDATE market_products SET stock=stock-$1 WHERE id=$2", input.Quantity, p.ID); e != nil {
		return domain.Order{}, e
	}
	o, e := scanOrder(tx.QueryRow(ctx, orderQuery+" WHERE o.id=$1", id))
	if e != nil {
		return o, e
	}
	return o, tx.Commit(ctx)
}
func (r *Repository) Advance(ctx context.Context, u domain.User, id int64, expected string) (domain.Order, error) {
	next := domain.NextStatus(expected)
	if next == "" {
		return domain.Order{}, domain.ErrInvalid
	}
	tag, e := r.DB.Exec(ctx, `UPDATE market_orders SET status=$1 WHERE id=$2 AND seller_id=$3 AND status=$4`, next, id, u.ID, expected)
	if e != nil {
		return domain.Order{}, e
	}
	if tag.RowsAffected() == 0 {
		return domain.Order{}, domain.ErrConflict
	}
	return scanOrder(r.DB.QueryRow(ctx, orderQuery+" WHERE o.id=$1 AND o.seller_id=$2", id, u.ID))
}
func (r *Repository) CreateOIDC(ctx context.Context, id, provider, nonce, role, proofHash string) error {
	_, e := r.DB.Exec(ctx, "INSERT INTO market_oidc_challenges(id,provider,nonce,role,proof_hash) VALUES($1,$2,$3,$4,$5)", id, provider, nonce, role, proofHash)
	return e
}
func (r *Repository) ConsumeOIDC(ctx context.Context, id, provider, nonce string) (string, error) {
	var role string
	e := r.DB.QueryRow(ctx, "DELETE FROM market_oidc_challenges WHERE id=$1 AND provider=$2 AND nonce=$3 AND expires_at>now() RETURNING role", id, provider, nonce).Scan(&role)
	if errors.Is(e, pgx.ErrNoRows) {
		e = domain.ErrUnauthorized
	}
	return role, e
}

func (r *Repository) CompleteApple(ctx context.Context, id, nonce, email string) error {
	tag, e := r.DB.Exec(ctx, "UPDATE market_oidc_challenges SET identity_email=$1 WHERE id=$2 AND nonce=$3 AND provider='Apple' AND expires_at>now() AND identity_email=''", email, id, nonce)
	if e == nil && tag.RowsAffected() == 0 {
		return domain.ErrUnauthorized
	}
	return e
}
func (r *Repository) ConsumeApple(ctx context.Context, id, proofHash string) (string, string, error) {
	var email, role string
	e := r.DB.QueryRow(ctx, "DELETE FROM market_oidc_challenges WHERE id=$1 AND proof_hash=$2 AND provider='Apple' AND expires_at>now() AND identity_email<>'' RETURNING identity_email,role", id, proofHash).Scan(&email, &role)
	if errors.Is(e, pgx.ErrNoRows) {
		e = domain.ErrNotFound
	}
	return email, role, e
}

// Cleanup removes expired authentication material and rate-limit identifiers.
func (r *Repository) Cleanup(ctx context.Context) error {
	for _, query := range []string{
		"DELETE FROM market_sessions WHERE expires_at<now()",
		"DELETE FROM market_challenges WHERE expires_at<now()",
		"DELETE FROM market_oidc_challenges WHERE expires_at<now()",
		"DELETE FROM market_limits WHERE until_at<now()",
	} {
		if _, e := r.DB.Exec(ctx, query); e != nil {
			return e
		}
	}
	return nil
}

// No upsert: registering an existing email cannot replace credentials or attach a session.
func (r *Repository) PasswordAuth(ctx context.Context, email, password, role, tokenHash string, register bool) (domain.User, error) {
	tx, err := r.DB.Begin(ctx)
	if err != nil {
		return domain.User{}, err
	}
	defer tx.Rollback(ctx)
	if register {
		hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
		if err != nil {
			return domain.User{}, err
		}
		tag, err := tx.Exec(ctx, `INSERT INTO market_users(email,role,provider,password_hash) VALUES($1,$2,'Email',$3) ON CONFLICT(email) DO NOTHING`, email, role, string(hash))
		if err != nil {
			return domain.User{}, err
		}
		if tag.RowsAffected() != 1 {
			return domain.User{}, domain.ErrUnauthorized
		}
	} else {
		var hash string
		err := tx.QueryRow(ctx, `SELECT COALESCE(password_hash,'') FROM market_users WHERE email=$1 AND NOT suspended FOR UPDATE`, email).Scan(&hash)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, err
		}
		missing := hash == ""
		if missing {
			// Keep unknown and non-password accounts on the same expensive verification path.
			hash = dummyPasswordHash
		}
		compareErr := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
		if compareErr != nil || missing {
			return domain.User{}, domain.ErrUnauthorized
		}
	}
	u, err := scanUser(tx.QueryRow(ctx, "SELECT "+userColumns+" FROM market_users WHERE email=$1 AND NOT suspended FOR UPDATE", email))
	if err != nil {
		return domain.User{}, err
	}
	if u.Role != role {
		return domain.User{}, domain.ErrUnauthorized
	}
	if _, err = tx.Exec(ctx, `INSERT INTO market_sessions(token_hash,user_id,expires_at) VALUES($1,$2,now()+interval '30 days')`, tokenHash, u.ID); err != nil {
		return domain.User{}, err
	}
	return u, tx.Commit(ctx)
}

var dummyPasswordHash = func() string {
	hash, err := bcrypt.GenerateFromPassword([]byte("invalid-account-dummy-credential"), 12)
	if err != nil {
		panic(err)
	}
	return string(hash)
}()

func (r *Repository) ProductsForUser(ctx context.Context, id string) ([]domain.Product, error) {
	rows, err := r.DB.Query(ctx, productQuery+` WHERE NOT p.hidden AND NOT u.suspended AND NOT EXISTS(SELECT 1 FROM market_blocks b WHERE b.user_id=$1 AND b.seller_id=p.seller_id) ORDER BY p.id DESC LIMIT 500`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := []domain.Product{}
	for rows.Next() {
		p, e := scanProduct(rows)
		if e != nil {
			return nil, e
		}
		list = append(list, p)
	}
	return list, rows.Err()
}
func (r *Repository) ReportProduct(ctx context.Context, user string, product int64, reason string) error {
	tag, err := r.DB.Exec(ctx, `INSERT INTO market_reports(reporter_id,product_id,reason) SELECT $1,id,$3 FROM market_products WHERE id=$2`, user, product, reason)
	if err == nil && tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return err
}
func (r *Repository) BlockSeller(ctx context.Context, user, seller string) error {
	if user == seller {
		return domain.ErrInvalid
	}
	tag, err := r.DB.Exec(ctx, `INSERT INTO market_blocks(user_id,seller_id) SELECT $1,id FROM market_users WHERE id::text=$2 AND role='Seller' ON CONFLICT(user_id,seller_id) DO UPDATE SET created_at=now()`, user, seller)
	if err == nil && tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return err
}
func (r *Repository) UnblockSellers(ctx context.Context, user string) error {
	_, err := r.DB.Exec(ctx, `DELETE FROM market_blocks WHERE user_id=$1`, user)
	return err
}
