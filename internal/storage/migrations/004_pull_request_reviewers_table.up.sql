CREATE TABLE pull_request_reviewers (
                                        pull_request_id VARCHAR(255) NOT NULL,
                                        user_id VARCHAR(255) NOT NULL,
                                        assigned_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                                        PRIMARY KEY (pull_request_id, user_id),
                                        FOREIGN KEY (pull_request_id) REFERENCES pull_requests(pull_request_id) ON DELETE CASCADE,
                                        FOREIGN KEY (user_id) REFERENCES users(user_id)
);

CREATE INDEX idx_pull_request_reviewers_user_id ON pull_request_reviewers(user_id);
CREATE INDEX idx_pull_request_reviewers_assigned_at ON pull_request_reviewers(assigned_at);