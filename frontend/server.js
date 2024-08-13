const express = require('express');
const path = require('path');

const app = express();
const port = 3000;
const ip = '192.168.10.25';

// Serve static files from 'public' directory
app.use(express.static(path.join(__dirname, 'public')));

// Redirect root to /index
app.get('/', (req, res) => {
    res.redirect('/index');
});

// Send HTML files
app.get('/index', (req, res) => {
    res.sendFile(path.join(__dirname, 'public', 'html', 'index.html'));
});

app.get('/chat', (req, res) => {
    res.sendFile(path.join(__dirname, 'public', 'html', 'chat.html'));
});

app.listen(port, ip, () => {
    console.log(`Frontend server running at http://${ip}:${port}`);
});