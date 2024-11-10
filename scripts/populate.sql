DO $$
DECLARE
    usernames TEXT[] := ARRAY['john_doe', 'jane_smith', 'alex_brown', 'emily_davis', 'michael_jones', 'chris_evans', 'maria_hernandez', 'raj_kumar', 'sophie_lee', 'liam_murphy', 'lucas_harris', 'isabella_gomez'];
    first_names TEXT[] := ARRAY['John', 'Jane', 'Alex', 'Emily', 'Michael', 'Chris', 'Maria', 'Raj', 'Sophie', 'Liam', 'Lucas', 'Isabella'];
    last_names TEXT[] := ARRAY['Doe', 'Smith', 'Brown', 'Davis', 'Jones', 'Evans', 'Hernandez', 'Kumar', 'Lee', 'Murphy', 'Harris', 'Gomez'];
    post_titles TEXT[] := ARRAY['Understanding PostgreSQL', 'Learning Go Programming', 'Frontend Tips', 'Backend Architecture', 'Database Optimization', 'API Design', 'Concurrency in Go', 'Testing in Software', 'Data Science Basics', 'Web Development Trends'];
    tags TEXT[] := ARRAY['tech', 'go', 'javascript', 'frontend', 'backend', 'data', 'design', 'tutorial', 'database', 'optimization'];
    content_snippets TEXT[] := ARRAY['This is a detailed guide on', 'Here are some best practices for', 'An in-depth look at', 'A brief introduction to', 'A comprehensive tutorial on', 'Exploring the concepts of', 'Key takeaways from', 'Tips and tricks for', 'Detailed insights into', 'An essential resource on'];
    user_id UUID;
    post_id UUID;
    follower_count INTEGER;

BEGIN
    -- Insert random users
    FOR i IN 1..100 LOOP
        INSERT INTO users (id, created_at, updated_at, username, email, password, first_name, last_name, is_deleted)
        VALUES (
            gen_random_uuid(),
            NOW() - (interval '1 day' * (i % 365)),
            NOW(),
            usernames[(1 + random() * (array_length(usernames, 1) - 1))::int] || '_' || i,
            'user' || i || '@example.com',
            'hashed_password',
            first_names[(1 + random() * (array_length(first_names, 1) - 1))::int],
            last_names[(1 + random() * (array_length(last_names, 1) - 1))::int],
            FALSE
        )
        RETURNING id INTO user_id;

        -- Each user creates some posts
        FOR j IN 1..(1 + random() * 20)::int LOOP
            INSERT INTO posts (id, created_at, updated_at, title, content, user_id, tags, is_deleted, version)
            VALUES (
                gen_random_uuid(),
                NOW() - (interval '1 day' * (j % 30)),
                NOW(),
                post_titles[(1 + random() * (array_length(post_titles, 1) - 1))::int] || ' #' || j,
                content_snippets[(1 + random() * (array_length(content_snippets, 1) - 1))::int] || ' topic.',
                user_id,
                ARRAY[
                    tags[(1 + random() * (array_length(tags, 1) - 1))::int],
                    tags[(1 + random() * (array_length(tags, 1) - 1))::int]
                ],
                FALSE,
                1
            )
            RETURNING id INTO post_id;

            -- Add random comments for each post
            FOR k IN 1..(random() * 50)::int LOOP
                INSERT INTO comments (id, post_id, user_id, created_at, updated_at, content)
                VALUES (
                    gen_random_uuid(),
                    post_id,
                    (SELECT id FROM users ORDER BY random() LIMIT 1),
                    NOW() - (interval '1 hour' * (k % 24)),
                    NOW(),
                    content_snippets[(1 + random() * (array_length(content_snippets, 1) - 1))::int] || ' comment'
                );
            END LOOP;
        END LOOP;

        -- Each user follows a random subset of other users
        follower_count := (1 + random() * 50)::int;
        FOR j IN 1..follower_count LOOP
            INSERT INTO followers (user_id, follower_id, created_at)
            VALUES (
                user_id,
                (SELECT id FROM users ORDER BY random() LIMIT 1),
                NOW() - (interval '1 day' * j)
            ) ON CONFLICT DO NOTHING;
        END LOOP;
    END LOOP;

END $$;
