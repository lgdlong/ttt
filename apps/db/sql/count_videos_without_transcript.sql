-- Find số video chưa có transcript nào
SELECT
    COUNT(*)
FROM
    videos v
WHERE
    NOT EXISTS (
        SELECT 1
        FROM transcript_segments ts
        WHERE ts.video_id = v.id
    );