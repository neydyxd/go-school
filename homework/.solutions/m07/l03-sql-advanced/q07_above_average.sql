SELECT title, price
FROM books
WHERE price > (SELECT avg(price) FROM books)
ORDER BY price DESC, title;
