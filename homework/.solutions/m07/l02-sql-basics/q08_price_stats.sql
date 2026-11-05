SELECT count(*) AS total,
       count(price) AS priced,
       round(avg(price), 2) AS avg_price
FROM books;
