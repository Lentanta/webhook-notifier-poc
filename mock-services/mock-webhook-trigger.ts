import express from 'npm:express';
import rateLimit from 'npm:express-rate-limit';
import postgres from 'npm:postgres';
import { crypto } from "jsr:@std/crypto/crypto";

import path from 'node:path';

const app = express();
const PORT = 3000;

const MAX_REQUEST = 20;

// Middleware
app.use(express.json());
app.use(express.urlencoded({ extended: true }));
app.use(express.static(path.join(import.meta.dirname, 'public')));

// PostgreSQL connection
const sql = postgres({
  host: 'localhost',
  port: 5432,
  database: 'events_db',
  username: 'postgres',
  password: 'postgres',
});

console.log('Connected to PostgreSQL');


// Mock customer POST API with a rate limit
const limiter = rateLimit({
  windowMs: 30 * 1000,
  max: MAX_REQUEST,
  message: {
    error: 'Too many requests, please try again later.',
    retryAfter: '1 minute'
  },
  standardHeaders: true,
  legacyHeaders: false,
  handler: (req, res) => {
    res.status(429).json({
      error: 'Too many requests',
      message: `You have exceeded the rate limit of ${MAX_REQUEST} requests per minute`,
      retryAfter: Math.ceil(req.rateLimit.resetTime / 1000)
    });
  }
});

// I'm adding this to make a fun random fail :)
function randomFalse(percentFalse) {
  return Math.random() * 100 >= percentFalse;
};

app.post('/customer-webhook', limiter, (req, res) => {
  const data = req.body;
  const isFail = randomFalse(30);

  console.log('Request received:', data);
  console.log(`Remaining requests: ${req.rateLimit.remaining}/${req.rateLimit.limit}`);

  if (isFail) {
    return res.status(503).json({ error: 'Fail to process the request' });
  }

  res.json({
    success: true,
    message: 'Data received successfully',
    receivedData: data,
    timestamp: new Date().toISOString(),
    rateLimit: {
      limit: req.rateLimit.limit,
      remaining: req.rateLimit.remaining,
      resetTime: new Date(req.rateLimit.resetTime).toISOString()
    }
  });
});

function generateSecureRandomString(length = 16) {
  const bytes = new Uint8Array(length);
  crypto.getRandomValues(bytes);
  return Array.from(bytes)
    .map(b => b.toString(16).padStart(2, '0'))
    .join('')
    .slice(0, length);
}

// HTML page with the form
app.get('/', (_, res) => {
  res.sendFile(path.join(__dirname, 'views', 'index.html'));
});

// API endpoint to create event in database
app.post('/api/send', async (req, res) => {
  const { message } = req.body;
  if (!message) return res.status(400).json({ error: 'No message' });

  try {
    const eventTime = new Date().toISOString();

    // Insert event into database
    const result = await sql`
      INSERT INTO events (event_name, event_time, payload, webhook_id)
      VALUES (
        ${message.event_name || 'test.event'},
        ${eventTime},
        ${sql.json(message)},
        ${"WH-" + generateSecureRandomString(8)}
      )
      RETURNING id
    `;

    res.json({
      success: true,
      message: 'Event inserted into database successfully',
      eventId: result[0].id,
      createdAt: eventTime
    });

  } catch (error) {
    console.error('Error inserting event:', error);
    res.status(500).json({ error: 'Failed to insert event into database' });
  }
});

app.listen(PORT, () => {
  console.log(`Server running on http://localhost:${PORT}`);
});
