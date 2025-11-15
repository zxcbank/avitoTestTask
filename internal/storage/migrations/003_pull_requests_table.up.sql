CREATE TABLE pull_requests (
                               pull_request_id VARCHAR(255) PRIMARY KEY,
                               pull_request_name VARCHAR(255) NOT NULL,
                               author_id VARCHAR(255) NOT NULL,
                               status VARCHAR(50) NOT NULL DEFAULT 'OPEN' CHECK (status IN ('OPEN', 'MERGED')),
                               created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                               merged_at TIMESTAMP NULL,
                               FOREIGN KEY (author_id) REFERENCES users(user_id)
);

CREATE INDEX idx_pull_requests_author_id ON pull_requests(author_id);
CREATE INDEX idx_pull_requests_status ON pull_requests(status);
CREATE INDEX idx_pull_requests_created_at ON pull_requests(created_at);