document.addEventListener('DOMContentLoaded', function() {
    const shortenBtn = document.getElementById('shortenBtn');
    const copyBtn = document.getElementById('copyBtn');
    const analyticsContainer = document.getElementById('resultCard');
    const shortUrlInput = document.getElementById('shortUrl');
    const analyticsLink = document.getElementById('analyticsLink');
    const errorContainer = document.getElementById('errorContainer');

    shortenBtn.addEventListener('click', async function() {
        const originalUrl = document.getElementById('originalUrl').value.trim();
        const customAlias = document.getElementById('customAlias').value.trim();
        
        // Валидация URL
        if (!originalUrl) {
            showError('Пожалуйста, введите URL');
            return;
        }

        // Проверяем, что URL начинается с http:// или https://
        if (!originalUrl.startsWith('http://') && !originalUrl.startsWith('https://')) {
            showError('URL должен начинаться с http:// или https://');
            return;
        }

        // Валидация custom alias (если указан)
        if (customAlias && !/^[a-zA-Z0-9]{3,20}$/.test(customAlias)) {
            showError('Alias может содержать только буквы и цифры (3-20 символов)');
            return;
        }

        try {
            const response = await fetch('/shorten', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    url: originalUrl,
                    custom: customAlias || ''
                })
            });

            const responseText = await response.text();
            
            if (!response.ok) {
                let errorMessage = `Ошибка: ${response.status}`;
                
                // Пытаемся распарсить JSON ошибки
                try {
                    const errorData = JSON.parse(responseText);
                    errorMessage = errorData.error || errorMessage;
                } catch (e) {
                    // Если не JSON, используем текст ответа
                    if (responseText) {
                        errorMessage = responseText;
                    }
                }
                
                // Специфичные ошибки
                if (errorMessage.includes('alias already exists') || response.status === 409) {
                    errorMessage = 'Этот alias уже занят. Пожалуйста, выберите другой.';
                } else if (errorMessage.includes('invalid alias') || errorMessage.includes('invalid url format')) {
                    errorMessage = 'Некорректный alias. Используйте только буквы, цифры (3-20 символов).';
                } else if (errorMessage.includes('invalid url')) {
                    errorMessage = 'Некорректный URL. Пожалуйста, введите корректный URL.';
                } else if (response.status === 500) {
                    errorMessage = 'Внутренняя ошибка сервера. Попробуйте позже.';
                }
                
                throw new Error(errorMessage);
            }

            // Успешный ответ
            const data = JSON.parse(responseText);
            const shortUrl = data.shortUrl || `${window.location.origin}/s/${data.alias}`;
            
            shortUrlInput.value = shortUrl;
            analyticsLink.href = `/analytics?alias=${data.alias}`;
            analyticsContainer.classList.remove('hidden');
            errorContainer.classList.add('hidden');
            
            // Очищаем форму
            document.getElementById('originalUrl').value = '';
            document.getElementById('customAlias').value = '';
            
        } catch (error) {
            showError(error.message);
        }
    });

    copyBtn.addEventListener('click', function() {
        shortUrlInput.select();
        
        // Используем современный API для копирования
        if (navigator.clipboard) {
            navigator.clipboard.writeText(shortUrlInput.value)
                .then(() => {
                    showCopyFeedback();
                })
                .catch(() => {
                    // Fallback для старых браузеров
                    document.execCommand('copy');
                    showCopyFeedback();
                });
        } else {
            // Fallback для старых браузеров
            document.execCommand('copy');
            showCopyFeedback();
        }
    });

    function showCopyFeedback() {
        const originalText = copyBtn.textContent;
        copyBtn.textContent = '✅ Скопировано!';
        setTimeout(() => {
            copyBtn.textContent = originalText;
        }, 2000);
    }

    function showError(message) {
        errorContainer.textContent = message;
        errorContainer.classList.remove('hidden');
        // Автоматически скрываем ошибку через 7 секунд
        setTimeout(() => {
            errorContainer.classList.add('hidden');
        }, 7000);
    }

    // Авто-фокус на поле URL при загрузке страницы
    document.getElementById('originalUrl').focus();
});