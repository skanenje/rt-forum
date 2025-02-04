document.addEventListener('DOMContentLoaded', () => {
    const app = document.getElementById('app');
    const authContainer = document.getElementById('auth-container');
    const forumContainer = document.getElementById('forum-container');
    
    // Form Toggle Logic
    const showLoginBtn = document.getElementById('show-login');
    const showRegisterBtn = document.getElementById('show-register');
    const loginForm = document.getElementById('login-form');
    const registerForm = document.getElementById('register-form');

    showLoginBtn.addEventListener('click', () => {
        loginForm.style.display = 'block';
        registerForm.style.display = 'none';
    });

    showRegisterBtn.addEventListener('click', () => {
        loginForm.style.display = 'none';
        registerForm.style.display = 'block';
    });
    const postForm = document.getElementById('post-form');
    if (postForm) {
        postForm.addEventListener('submit', async (e) => {
            e.preventDefault(); // Prevent form from submitting normally
            
            const formData = new FormData(postForm);
            const postData = {
                title: formData.get('title'),
                content: formData.get('content'),
                category: formData.get('category')
            };

            try {
                const response = await fetch('/api/posts', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json'
                    },
                    body: JSON.stringify(postData)
                });

                if (response.ok) {
                    // Clear form
                    postForm.reset();
                    // Reload posts
                    loadPosts();
                } else if (response.status === 401) {
                    // Session expired - show login form
                    document.getElementById('auth-container').style.display = 'flex';
                    document.getElementById('forum-container').style.display = 'none';
                } else {
                    alert('Error creating post. Please try again.');
                }
            } catch (error) {
                console.error('Error creating post:', error);
                alert('Error creating post. Please try again.');
            }
        });
        }
    // Authentication Handling
    async function handleRegistration(e) {
        e.preventDefault();
        const formData = new FormData(e.target);
        const data = Object.fromEntries(formData.entries());
        const age = Number(data.age);
        if (isNaN(age) || age < 13) {
            document.getElementById('register-errors').innerHTML = 
                '<p>Please enter a valid age (must be 13 or older)</p>';
            return;
        }
        data.age = age;
        const registerErrors = document.getElementById('register-errors');
        // Add this before the fetch call
        console.log('Sending data:', JSON.stringify(data, null, 2));
        try {
            const response = await fetch('/api/register', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(data)
            });
    
            const result = await response.json();
            
            if (response.ok) {
                alert('Registration successful! Please log in.');
                loginForm.style.display = 'block';
                registerForm.style.display = 'none';
                registerErrors.innerHTML = '';
            } else {
                registerErrors.innerHTML = result.errors.map(err => `<p>${err}</p>`).join('');
            }
        } catch (error) {
            console.error('Registration error:', error);
            registerErrors.innerHTML = '<p>Network error. Please try again.</p>';
        }
    }
    async function handleLogin(e) {
        e.preventDefault();
        const formData = new FormData(e.target);
        const data = Object.fromEntries(formData.entries());
        const loginErrors = document.getElementById('login-errors');
        
        try {
            const response = await fetch('/api/login', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(data)
            });

            const result = await response.json();
            
            if (result.success) {
                // Store the session token
                localStorage.setItem('sessionToken', result.session_token);
                
                // Hide auth, show forum
                authContainer.style.display = 'none';
                forumContainer.style.display = 'grid';
                
                // Update user info
                document.getElementById('logged-user').textContent = `Welcome, ${result.nickname}`;
                
                // Initialize WebSocket and other features
                initializeForumFeatures(result);
            } else {
                loginErrors.innerHTML = `<p>${
                    result.error_type === 'user_not_found' 
                    ? 'User not found' 
                    : 'Incorrect password'
                }</p>`;
            }
        } catch (error) {
            console.error('Login error:', error);
            loginErrors.innerHTML = '<p>Network error. Please try again.</p>';
        }
    }
    function initializeForumFeatures(userInfo) {
        // WebSocket initialization
        const socket = new WebSocket('ws://localhost:8080/ws/chat');
        
        // Posts loading
        loadPosts();
        
        // Online users tracking
        trackOnlineUsers();
        
        // Private messaging setup
        setupPrivateMessaging(socket, userInfo);
    }

    function loadPosts() {
        fetch('/api/posts')
            .then(res => res.json())
            .then(posts => {
                const postsContainer = document.getElementById('posts-container');
                postsContainer.innerHTML = posts.map(post => `
                    <div class="post">
                        <h3>${post.title}</h3>
                        <p>${post.content}</p>
                        <small>Category: ${post.category}</small>
                    </div>
                `).join('');
            });
    }

    function trackOnlineUsers() {
        // Periodic online users update
        setInterval(() => {
            fetch('/api/online-users')
                .then(res => res.json())
                .then(users => {
                    const onlineUsersContainer = document.getElementById('online-users');
                    onlineUsersContainer.innerHTML = users.map(user => `
                        <div class="user" data-user-id="${user.id}">
                            ${user.nickname}
                        </div>
                    `).join('');
                });
        }, 60000);
    }

    function setupPrivateMessaging(socket, userInfo) {
        // Private message sending logic
        const messageInput = document.getElementById('message-input');
        const sendMessageBtn = document.getElementById('send-message-btn');

        sendMessageBtn.addEventListener('click', () => {
            const receiverId = document.querySelector('.active-chat').dataset.userId;
            const message = messageInput.value;

            socket.send(JSON.stringify({
                sender_id: userInfo.user_id,
                receiver_id: receiverId,
                content: message
            }));

            messageInput.value = '';
        });

        // Handle incoming messages
        socket.onmessage = (event) => {
            const message = JSON.parse(event.data);
            displayMessage(message);
        };
    }

    // Event Listeners
    document.getElementById('register-form').addEventListener('submit', handleRegistration);
    document.getElementById('login-form').addEventListener('submit', handleLogin);
    document.getElementById('logout-btn').addEventListener('click', () => {
        // Clear session and redirect to login
        document.cookie = 'session=; Path=/; Expires=Thu, 01 Jan 1970 00:00:01 GMT;';
        authContainer.style.display = 'flex';
        forumContainer.style.display = 'none';
    });
});