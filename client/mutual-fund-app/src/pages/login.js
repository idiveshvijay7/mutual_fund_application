import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import axios from 'axios';

function LoginPage() {
    const [phoneNumber, setPhoneNumber] = useState('');
    const navigate = useNavigate();

    const handlePhoneNumberChange = (event) => {
        setPhoneNumber(event.target.value);
    };

    const handleLogin = async () => {
        if (!phoneNumber.trim()) {
            alert('Please enter a phone number.');
            return;
        }

        try {
            const response = await axios.post('http://localhost:8081/login',
                { phoneNumber },
                {
                    headers: {
                        'Content-Type': 'application/json',
                    }
                }
            );

            if (response.status === 200) {
                // Store the phone number in local storage
                localStorage.setItem('phoneNumber', phoneNumber);

                // Redirect to home page upon successful login
                navigate('/home');
            } else {
                alert('Login failed. Please try again.');
            }
        } catch (error) {
            console.error('Error logging in:', error);
            alert('User Not exist, Please sign up');
        }
    };

    const handleSignup = async () => {
        if (!phoneNumber.trim()) {
            alert('Please enter a phone number.');
            return;
        }

        try {
            const response = await axios.post('http://localhost:8081/signup', { phoneNumber });

            if (response.status === 201) {
                // Store the phone number in local storage
                localStorage.setItem('phoneNumber', phoneNumber);

                // Redirect to home page upon successful signup
                navigate('/home');
            } else {
                // Handle signup failure
                alert('Signup failed. Please try again.');
            }
        } catch (error) {
            console.error('Error signing up:', error);
            alert('Internal Error retry');
        }
    };

    return (
        <div style={styles.container}>
            <div style={styles.box}>
                <h2 style={styles.heading}>Phone Number</h2>
                <input
                    type="text"
                    value={phoneNumber}
                    onChange={handlePhoneNumberChange}
                    placeholder="Enter your phone number"
                    style={styles.input}
                />
                <div style={styles.buttonContainer}>
                    <button onClick={handleLogin} style={styles.button}>Login</button>
                    <button onClick={handleSignup} style={styles.button}>Signup</button>
                </div>
            </div>
        </div>
    );
}

const styles = {
    container: {
        display: 'flex',
        justifyContent: 'center',
        alignItems: 'center',
        height: '100vh',
    },
    box: {
        width: '300px',
        padding: '20px',
        border: '1px solid #ccc',
        borderRadius: '5px',
        textAlign: 'center',
    },
    heading: {
        margin: '0',
    },
    input: {
        width: '100%',
        marginTop: '20px',
        padding: '10px',
        fontSize: '16px',
        border: '1px solid #ccc',
        borderRadius: '5px',
        boxSizing: 'border-box',
    },
    buttonContainer: {
        marginTop: '20px',
    },
    button: {
        margin: '0 10px',
        padding: '10px 20px',
        fontSize: '16px',
        cursor: 'pointer',
        backgroundColor: '#007bff',
        color: '#fff',
        border: 'none',
        borderRadius: '5px',
        outline: 'none',
    },
};

export default LoginPage;
