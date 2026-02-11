import express from 'npm:express';
import amqp from 'npm:amqplib';
import rateLimit from 'npm:express-rate-limit';

import path from 'node:path';

const app = express();
const PORT = 3000;

const MAX_REQUEST = 20;

// Middleware
app.use(express.json());
app.use(express.urlencoded({ extended: true }));
app.use(express.static(path.join(import.meta.dirname, 'public')));

// RabbitMQ connection
let channel: any;
const QUEUE_NAME = 'webhook_queue';

async function connectRabbitMQ() {
  try {
    const connection = await amqp.connect('amqp://localhost');
    channel = await connection.createChannel();
    await channel.assertQueue(QUEUE_NAME, { durable: false });
    console.log('Connected to RabbitMQ');

  } catch (error) {
    console.error('RabbitMQ connection error:', error);
  }
}
connectRabbitMQ();


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

// HTML page with the form
app.get('/', (_, res) => {
  res.sendFile(path.join(__dirname, 'views', 'index.html'));
});

// API endpoint for send message to queue
app.post('/api/send', async (req, res) => {
  const { message } = req.body;

  if (!message) return res.status(400).json({ error: 'No message' });
  if (!channel) return res.status(503).json({ error: 'No channel' });

  try {
    const messageData = {
      content: message,
      timestamp: new Date().toISOString()
    };

    channel.sendToQueue(
      QUEUE_NAME,
      Buffer.from(JSON.stringify(messageData)),
      { persistent: true }
    );

    console.log('Message sent to queue:', message);

    res.json({
      success: true,
      message: 'Message sent to queue successfully',
      queuedAt: messageData.timestamp
    });

  } catch (error) {
    console.error('Error sending message:', error);
    res.status(500).json({ error: 'Failed to send message to queue' });
  }
});

app.listen(PORT, () => {
  console.log(`Server running on http://localhost:${PORT}`);
});
