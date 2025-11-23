document.addEventListener('DOMContentLoaded', function() {
    const shortenBtn = document.getElementById('shortenBtn');
    const copyBtn = document.getElementById('copyBtn');
    const analyticsContainer = document.getElementById('resultCard');
    const shortUrlInput = document.getElementById('shortUrl');
    const analyticsLink = document.getElementById('analyticsLink');
    const errorContainer = document.getElementById('errorContainer');

    shortenBtn.addEventListener('click', async function() {
        const originalUrl = document.getElementById('originalUrl').value;
        const customAlias = document.getElementById('customAlias').value;
        
        if (!originalUrl) {
            showError('Пожалуйста, введите URL');
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
                    custom: customAlias
                })
            });

            if (!response.ok) {
                const errorData = await response.json();
                throw new Error(errorData.error || 'Ошибка при создании короткой ссылки');
            }

            const data = await response.json();
            const shortUrl = `${window.location.origin}/s/${data.alias}`;
            
            shortUrlInput.value = shortUrl;
            analyticsLink.href = `/analytics?alias=${data.alias}`;
            analyticsContainer.classList.remove('hidden');
            errorContainer.classList.add('hidden');
            
            document.getElementById('originalUrl').value = '';
            document.getElementById('customAlias').value = '';
            
        } catch (error) {
            showError(error.message);
        }
    });

    copyBtn.addEventListener('click', function() {
        shortUrlInput.select();
        document.execCommand('copy');
        copyBtn.textContent = '✅ Скопировано!';
        setTimeout(() => {
            copyBtn.textContent = '📋 Копировать';
        }, 2000);
    });

    function showError(message) {
        errorContainer.textContent = message;
        errorContainer.classList.remove('hidden');
        setTimeout(() => {
            errorContainer.classList.add('hidden');
        }, 5000);
    }
});