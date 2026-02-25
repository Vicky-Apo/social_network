/* =========================
   DROP POST CATEGORIES
   ========================= */

DROP TRIGGER IF EXISTS trg_prevent_group_post_categories ON post_categories;
DROP FUNCTION IF EXISTS prevent_group_post_categories();
DROP TABLE IF EXISTS post_categories;
DROP TABLE IF EXISTS categories;
