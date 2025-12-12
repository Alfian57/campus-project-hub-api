-- +migrate Up
CREATE TABLE project_images (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    image_url TEXT NOT NULL,
    sort_order INTEGER DEFAULT 0
);

CREATE INDEX idx_project_images_project_id ON project_images(project_id);

CREATE TABLE project_likes (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, project_id)
);

CREATE INDEX idx_project_likes_project_id ON project_likes(project_id);

-- +migrate Down
DROP TABLE IF EXISTS project_likes;
DROP TABLE IF EXISTS project_images;
