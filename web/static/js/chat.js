// Chat application logic
document.addEventListener('alpine:init', () => {
    Alpine.data('chatApp', () => ({
        messages: [],
        newMessage: '',
        isLoading: false,
        isTyping: false,
        userId: 'user-' + Date.now(),
        ws: null,
        chatId: new URLSearchParams(window.location.search).get('id') || 'new',
        messagesLoading: true,
        loadError: null,
        reconnectAttempts: 0,

        init() {
            this.loadMessages();
            this.connectWebSocket();
        },

        loadMessages() {
            // Skip loading messages for new chats
            if (this.chatId === 'new') {
                this.messages = [];
                this.messagesLoading = false;
                return;
            }
            this.messagesLoading = true;
            this.loadError = null;
            // Fetch chat history from API
            fetch(`/api/v1/chat/${this.chatId}`)
                .then(response => {
                    if (!response.ok) {
                        throw new Error(`Failed to load messages: ${response.status}`);
                    }
                    return response.json();
                })
                .then(data => {
                    if (data.messages && Array.isArray(data.messages)) {
                        this.messages = data.messages.map(msg => ({
                            role: msg.role,
                            content: msg.content,
                            timestamp: new Date(msg.timestamp)
                        }));
                    } else {
                        this.messages = [];
                    }
                    this.messagesLoading = false;
                    // Scroll to bottom after messages are rendered
                    this.$nextTick(() => {
                        this.scrollToBottom();
                    });
                })
                .catch(error => {
                    console.error('Error loading messages:', error);
                    this.loadError = 'Failed to load chat history. Please try again.';
                    this.messagesLoading = false;
                });
        },

        connectWebSocket() {
            if (this.ws && this.ws.readyState !== WebSocket.CLOSED) {
                return; // Already connected or connecting
            }

            const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
            this.ws = new WebSocket(`${protocol}//${window.location.host}/ws/${this.userId}`);

            this.ws.onopen = () => {
                console.log('Connected to WebSocket');
                // Reset reconnection attempts
                this.reconnectAttempts = 0;
            };

            this.ws.onmessage = (event) => {
                const message = JSON.parse(event.data);
                if (message.type === 'chat_message') {
                    if (message.content) {
                        // Handle streaming chunks
                        const chatMessage = message.content;
                        // If this is the first chunk of a response, create a new message
                        if (this.isTyping) {
                            this.isTyping = false;
                            this.messages.push({
                                role: 'assistant',
                                content: chatMessage.content || '',
                                timestamp: new Date(chatMessage.timestamp)
                            });
                        } else if (chatMessage.content) {
                            // Append to the last message for streaming updates
                            const lastMessage = this.messages[this.messages.length - 1];
                            if (lastMessage && lastMessage.role === 'assistant') {
                                lastMessage.content += chatMessage.content;
                            }
                        }
                        this.scrollToBottom();
                    } else if (message.metadata && message.metadata.complete) {
                        // Message is complete, can update UI if needed
                        console.log('Message complete, processing time:', message.metadata.processingTime);
                        this.isLoading = false;
                    }
                } else if (message.type === 'typing') {
                    this.isTyping = true;
                    this.scrollToBottom();
                } else if (message.type === 'pong') {
                    // Received pong from server
                }
            };

            this.ws.onclose = (event) => {
                console.log(`WebSocket closed: ${event.code} ${event.reason}`);
                // Implement exponential backoff for reconnection
                const delay = Math.min(1000 * Math.pow(1.5, this.reconnectAttempts), 10000);
                this.reconnectAttempts++;
                console.log(`Reconnecting in ${delay}ms (attempt ${this.reconnectAttempts})`);
                setTimeout(() => this.connectWebSocket(), delay);
            };

            this.ws.onerror = (error) => {
                console.error('WebSocket error:', error);
            };

            // Handle ping/pong
            setInterval(() => {
                if (this.ws && this.ws.readyState === WebSocket.OPEN) {
                    this.ws.send(JSON.stringify({ type: "ping" }));
                }
            }, 30000);
        },

        sendMessage() {
            if (!this.newMessage.trim() || this.isLoading) return;

            const message = {
                role: 'user',
                content: this.newMessage.trim(),
                timestamp: new Date()
            };

            this.messages.push(message);
            this.scrollToBottom();

            // Clear input field
            const messageText = this.newMessage.trim();
            this.newMessage = '';

            // If this is a new chat, we need to create it first
            let apiRequest = Promise.resolve(this.chatId);

            if (this.chatId === 'new') {
                // Create a new chat
                apiRequest = fetch('/api/v1/chat', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                    body: JSON.stringify({
                        title: messageText.substring(0, 30) + (messageText.length > 30 ? '...' : '')
                    })
                })
                    .then(response => {
                        if (!response.ok) {
                            throw new Error('Failed to create chat');
                        }
                        return response.json();
                    })
                    .then(data => {
                        // Update URL with new chat ID without page reload
                        const url = new URL(window.location);
                        url.searchParams.set('id', data.id);
                        window.history.pushState({}, '', url);
                        // Update chatId
                        this.chatId = data.id;
                        return data.id;
                    });
            }

            // Send message via WebSocket after chat is created (if needed)
            apiRequest
                .then(chatId => {
                    this.isLoading = true;
                    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
                        this.ws.send(JSON.stringify({
                            type: 'chat_message',
                            content: {
                                chatId: chatId,
                                content: messageText,
                                role: 'user',
                                timestamp: message.timestamp
                            }
                        }));
                        // Show typing indicator
                        setTimeout(() => {
                            this.isTyping = true;
                            this.scrollToBottom();
                        }, 300);
                    } else {
                        throw new Error('WebSocket not connected');
                    }
                })
                .catch(error => {
                    console.error('Error sending message:', error);
                    // Add error message to chat
                    this.messages.push({
                        role: 'system',
                        content: 'Failed to send message. Please try again.',
                        timestamp: new Date()
                    });
                    this.isLoading = false;
                    this.scrollToBottom();
                });
        },

        scrollToBottom() {
            setTimeout(() => {
                const scrollAnchor = document.getElementById('scroll-anchor');
                if (scrollAnchor) {
                    scrollAnchor.scrollIntoView({ behavior: 'smooth' });
                }
            }, 100);
        },

        formatTime(timestamp) {
            if (!timestamp) return '';
            const date = new Date(timestamp);
            return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
        },

        formatMessage(content) {
            // Simple markdown-like formatting
            if (!content) return '';
            // Format code blocks
            content = content.replace(/```(\w+)?\n([\s\S]*?)\n```/g, '<div class="code-block"><pre><code>$2</code></pre></div>');
            // Format inline code
            content = content.replace(/`([^`]+)`/g, '<code class="bg-gray-700 dark:bg-gray-700 px-1 rounded text-gray-200">$1</code>');
            // Convert line breaks to <br>
            content = content.replace(/\n/g, '<br>');
            return content;
        },

        autoGrow(element) {
            element.style.height = 'auto';
            element.style.height = (element.scrollHeight) + 'px';
            // Limit to 5 rows
            const lineHeight = parseInt(getComputedStyle(element).lineHeight);
            const maxHeight = lineHeight * 5;
            if (element.scrollHeight > maxHeight) {
                element.style.height = maxHeight + 'px';
                element.style.overflowY = 'auto';
            } else {
                element.style.overflowY = 'hidden';
            }
        }
    }))
})