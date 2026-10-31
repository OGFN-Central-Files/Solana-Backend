package utilities

import "fmt"

func GenerateSuccessHTML(userID, avatar, username, accountID, token, displayName string) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>Solaris | Success</title>
	<link rel="icon" href="https://cdn.itztiva.com/Solarislogo.png">
	<meta content="#9837F9" data-react-helmet="true" name="theme-color" />
	<link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&display=swap" rel="stylesheet">
	<style>
		* {
			margin: 0;
			padding: 0;
			box-sizing: border-box;
		}

		body {
			display: flex;
			justify-content: center;
			align-items: center;
			min-height: 100vh;
			background: linear-gradient(135deg, #1a1c2e, #2b1155);
			font-family: 'Inter', sans-serif;
			color: #fff;
			position: relative;
			overflow: hidden;
		}

		.background-shapes {
			position: absolute;
			width: 100%%;
			height: 100%%;
			overflow: hidden;
			z-index: 0;
		}

		.shape {
			position: absolute;
			background: rgba(255, 255, 255, 0.03);
			border-radius: 50%%;
			animation: float 20s infinite linear;
		}

		@keyframes float {
			0%% { transform: translateY(0) rotate(0deg); }
			100%% { transform: translateY(-100vh) rotate(360deg); }
		}

		.auth-container {
			position: relative;
			z-index: 1;
			width: 90%%;
			max-width: 400px;
			padding: 1.0rem;
			background: rgba(255, 255, 255, 0.05);
			backdrop-filter: blur(10px);
			border-radius: 20px;
			border: 1px solid rgba(255, 255, 255, 0.1);
			box-shadow: 0 20px 40px rgba(0, 0, 0, 0.2);
			animation: container-appear 0.6s ease-out;
		}

		@keyframes container-appear {
			from {
				opacity: 0;
				transform: translateY(20px);
			}
			to {
				opacity: 1;
				transform: translateY(0);
			}
		}

		.profile-section {
			display: flex;
			align-items: center;
			gap: 20px;
			margin-bottom: 2rem;
		}

		.profile-img {
			width: 64px;
			height: 64px;
			border-radius: 16px;
			box-shadow: 0 4px 8px rgba(0, 0, 0, 0.2);
			transition: transform 0.3s ease;
		}

		.profile-img:hover {
			transform: scale(1.05);
		}

		.user-info {
			flex-grow: 1;
		}

		.username {
			font-size: 1.5rem;
			font-weight: 700;
			margin-bottom: 0.5rem;
			background: linear-gradient(45deg, #fff, #b19fff);
			-webkit-background-clip: text;
			-webkit-text-fill-color: transparent;
		}

		.user-id {
			color: rgba(255, 255, 255, 0.6);
			font-size: 0.875rem;
			font-family: monospace;
		}

		.login-button {
			display: block;
			width: 100%%;
			padding: 1rem;
			background: linear-gradient(45deg, #444579, rgba(99, 101, 241, 0.637));
			color: white;
			text-decoration: none;
			border-radius: 12px;
			font-weight: 600;
			text-align: center;
			transition: all 0.3s ease;
			border: none;
			cursor: pointer;
			position: relative;
			overflow: hidden;
		}

		.login-button:hover {
			transform: translateY(-2px);
			box-shadow: 0 8px 16px rgba(152, 55, 249, 0.2);
		}

		.login-button:active {
			transform: translateY(0);
		}

		.login-button.success {
			background: linear-gradient(45deg, #10B981, #059669);
		}

		.success-icon {
			display: inline-block;
			margin-right: 8px;
			animation: check-appear 0.5s ease-out;
		}

		@keyframes check-appear {
			from {
				transform: scale(0);
				opacity: 0;
			}
			to {
				transform: scale(1);
				opacity: 1;
			}
		}

		@media (max-width: 480px) {
			.auth-container {
				padding: 1.5rem;
			}

			.profile-img {
				width: 48px;
				height: 48px;
			}

			.username {
				font-size: 1.25rem;
			}
		}
	</style>
</head>
<body>
	<div class="background-shapes">
		<!-- these get added by the js code so ignore this -->
	</div>

	<div class="auth-container">
		<div class="profile-section">
			<img class="profile-img" src="https://cdn.discordapp.com/avatars/%s/%s" alt="Profile">
			<div class="user-info">
				<div class="username">%s</div>
				<div class="user-id">%s</div>
			</div>
		</div>
		<a href="solaris://%s" class="login-button">
			Continue to Solaris as <strong>%s</strong>
		</a>
	</div>

<script>
		const shapes = document.querySelector('.background-shapes');
		for (let i = 0; i < 15; i++) {
			const shape = document.createElement('div');
			shape.className = 'shape';
			const size = Math.random() * 100 + 50;
			shape.style.width = size + "px";
			shape.style.height = size + "px";
			shape.style.left = Math.random() * 100 + "%%"; 
			shape.style.top = Math.random() * 100 + "%%";  
			shape.style.animationDelay = Math.random() * 20 + "s";
			shapes.appendChild(shape);
		}

		document.querySelector('.login-button').addEventListener('click', function(e) {
			const button = this;
			button.classList.add('success');
			button.innerHTML = '<span class="success-icon">✓</span> Logged in successfully';
			button.style.pointerEvents = 'none';
		});
	</script>
</body>
</html>`, userID, avatar, username, accountID, token, displayName)
}
