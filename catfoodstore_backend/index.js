require("dotenv").config();
const express = require("express");
const { Pool } = require("pg");

const app = express();
app.use(express.json());

// ------------------------------
// PostgreSQL Connection
// ------------------------------
const pool = new Pool({
  host: process.env.DB_HOST,
  port: process.env.DB_PORT,
  user: process.env.DB_USER,
  password: process.env.DB_PASSWORD,
  database: process.env.DB_NAME,
});

// ------------------------------
// TEST connection
// ------------------------------
app.get("/api/test-db", async (req, res) => {
  try {
    const result = await pool.query("SELECT NOW()");
    res.json({ db: "connected", time: result.rows[0].now });
  } catch (err) {
    res.status(500).json({ error: err.message });
  }
});

// ============================================================
// 1) GET ALL PRODUCTS
// ============================================================
app.get("/api/products", async (req, res) => {
  try {
    const result = await pool.query("SELECT * FROM products ORDER BY id ASC");
    res.json(result.rows);
  } catch (err) {
    res.status(500).json({ error: err.message });
  }
});

// ============================================================
// 2) GET PRODUCT BY ID
// ============================================================
app.get("/api/products/:id", async (req, res) => {
  const { id } = req.params;

  try {
    const result = await pool.query(
      "SELECT * FROM products WHERE id = $1",
      [id]
    );

    if (result.rows.length === 0) {
      return res.status(404).json({ error: "Product not found" });
    }

    res.json(result.rows[0]);
  } catch (err) {
    res.status(500).json({ error: err.message });
  }
});

// ============================================================
// 3) CREATE PRODUCT
// ============================================================
app.post("/api/products", async (req, res) => {
  const {
    name,
    description,
    price,
    weight,
    age_group,
    breed_type,
    category,
    image_url,
  } = req.body;

  try {
    const result = await pool.query(
      `INSERT INTO products 
      (name, description, price, weight, age_group, breed_type, category, image_url) 
      VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
      RETURNING *`,
      [
        name,
        description,
        price,
        weight,
        age_group,
        breed_type,
        category,
        image_url,
      ]
    );

    res.status(201).json(result.rows[0]);
  } catch (err) {
    res.status(500).json({ error: err.message });
  }
});

// ============================================================
// 4) UPDATE PRODUCT
// ============================================================
app.put("/api/products/:id", async (req, res) => {
  const { id } = req.params;

  const {
    name,
    description,
    price,
    weight,
    age_group,
    breed_type,
    category,
    image_url,
  } = req.body;

  try {
    const result = await pool.query(
      `UPDATE products SET
        name = $1,
        description = $2,
        price = $3,
        weight = $4,
        age_group = $5,
        breed_type = $6,
        category = $7,
        image_url = $8,
        updated_at = NOW()
      WHERE id = $9
      RETURNING *`,
      [
        name,
        description,
        price,
        weight,
        age_group,
        breed_type,
        category,
        image_url,
        id
      ]
    );

    if (result.rows.length === 0) {
      return res.status(404).json({ error: "Product not found" });
    }

    res.json(result.rows[0]);
  } catch (err) {
    res.status(500).json({ error: err.message });
  }
});

// ============================================================
// 5) DELETE PRODUCT
// ============================================================
app.delete("/api/products/:id", async (req, res) => {
  const { id } = req.params;

  try {
    const result = await pool.query(
      "DELETE FROM products WHERE id = $1 RETURNING *",
      [id]
    );

    if (result.rows.length === 0) {
      return res.status(404).json({ error: "Product not found" });
    }

    res.json({ message: "Deleted successfully", product: result.rows[0] });
  } catch (err) {
    res.status(500).json({ error: err.message });
  }
});

// ------------------------------
// SERVER START
// ------------------------------
const PORT = process.env.PORT;
app.listen(PORT, () => console.log(`Backend running on port ${PORT}`));
