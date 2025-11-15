CREATE TABLE users (
                       user_id VARCHAR(255) PRIMARY KEY,
                       username VARCHAR(255) NOT NULL,
                       team_name VARCHAR(255) NOT NULL,
                       is_active BOOLEAN NOT NULL DEFAULT true,
                       FOREIGN KEY (team_name) REFERENCES teams(team_name) ON DELETE CASCADE
);

CREATE INDEX idx_users_team_name ON users(team_name);
CREATE INDEX idx_users_is_active ON users(is_active);