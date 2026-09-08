-- =============================================================================
-- usercenter schema
--   PostgreSQL 14+
--   与 internal/model/model.go 一一对应 (uc_* 表前缀)
--   与 GORM AutoMigrate 结果一致; 若两者差异以本文件为准
-- =============================================================================

-- -----------------------------------------------------------------------------
-- uc_users
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS uc_users (
    id              VARCHAR(36)   PRIMARY KEY,
    username        VARCHAR(64)   NOT NULL,
    email           VARCHAR(128),
    phone           VARCHAR(32),
    password_hash   VARCHAR(255)  NOT NULL,
    nickname        VARCHAR(64),
    avatar          VARCHAR(255),
    disabled        BOOLEAN       NOT NULL DEFAULT FALSE,
    created_at      TIMESTAMP     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at      TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_uc_users_username ON uc_users (username);
CREATE UNIQUE INDEX IF NOT EXISTS idx_uc_users_email    ON uc_users (email)    WHERE email IS NOT NULL;
CREATE INDEX        IF NOT EXISTS idx_uc_users_phone    ON uc_users (phone)    WHERE phone IS NOT NULL;
CREATE INDEX        IF NOT EXISTS idx_uc_users_deleted  ON uc_users (deleted_at);

COMMENT ON TABLE  uc_users                IS '用户主表';
COMMENT ON COLUMN uc_users.id              IS '用户 ID (UUID)';
COMMENT ON COLUMN uc_users.username        IS '登录名';
COMMENT ON COLUMN uc_users.email           IS '邮箱';
COMMENT ON COLUMN uc_users.phone           IS '手机号';
COMMENT ON COLUMN uc_users.password_hash   IS 'bcrypt 哈希';
COMMENT ON COLUMN uc_users.nickname        IS '昵称';
COMMENT ON COLUMN uc_users.avatar          IS '头像 URL';
COMMENT ON COLUMN uc_users.disabled        IS '是否禁用';
COMMENT ON COLUMN uc_users.deleted_at      IS '软删除时间';

-- -----------------------------------------------------------------------------
-- uc_refresh_tokens
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS uc_refresh_tokens (
    id          BIGSERIAL    PRIMARY KEY,
    user_id     VARCHAR(36)  NOT NULL,
    token_hash  VARCHAR(255) NOT NULL,
    expires_at  TIMESTAMP    NOT NULL,
    revoked     BOOLEAN      NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_uc_refresh_tokens_hash     ON uc_refresh_tokens (token_hash);
CREATE INDEX        IF NOT EXISTS idx_uc_refresh_tokens_user_id  ON uc_refresh_tokens (user_id);
CREATE INDEX        IF NOT EXISTS idx_uc_refresh_tokens_expires  ON uc_refresh_tokens (expires_at);

COMMENT ON TABLE  uc_refresh_tokens             IS '刷新令牌表';
COMMENT ON COLUMN uc_refresh_tokens.token_hash  IS 'refresh token 的 SHA-256 摘要';

-- -----------------------------------------------------------------------------
-- uc_oauth_clients
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS uc_oauth_clients (
    client_id      VARCHAR(64)  PRIMARY KEY,
    client_secret  VARCHAR(255) NOT NULL,
    name           VARCHAR(128),
    redirect_uris  TEXT,
    scopes         VARCHAR(255),
    created_at     TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at     TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON TABLE  uc_oauth_clients              IS 'OIDC 客户端表';
COMMENT ON COLUMN uc_oauth_clients.redirect_uris IS '允许的回调地址, 多值以换行分隔';
COMMENT ON COLUMN uc_oauth_clients.scopes        IS '允许的 scope, 空格分隔';

-- -----------------------------------------------------------------------------
-- uc_authorization_codes
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS uc_authorization_codes (
    code          VARCHAR(128) PRIMARY KEY,
    client_id     VARCHAR(64)  NOT NULL,
    user_id       VARCHAR(36)  NOT NULL,
    redirect_uri  VARCHAR(512),
    scope         VARCHAR(255),
    expires_at    TIMESTAMP    NOT NULL,
    used          BOOLEAN      NOT NULL DEFAULT FALSE,
    created_at    TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_uc_authcodes_client_id ON uc_authorization_codes (client_id);
CREATE INDEX IF NOT EXISTS idx_uc_authcodes_user_id   ON uc_authorization_codes (user_id);
CREATE INDEX IF NOT EXISTS idx_uc_authcodes_expires   ON uc_authorization_codes (expires_at);

COMMENT ON TABLE uc_authorization_codes IS 'OIDC 授权码 (一次性, 短期有效)';
