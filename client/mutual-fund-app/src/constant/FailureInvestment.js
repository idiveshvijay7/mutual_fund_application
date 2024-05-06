import React from 'react';
import { useNavigate } from 'react-router-dom';

const FailurePage = () => {
  const navigate = useNavigate();

  const handleRetry = () => {
    // Navigate to the Select Strategy page
    navigate('/select-strategy');
  };

  const handleContinue = () => {
    // Navigate to the Home page
    navigate('/home');
  };

  return (
    <div style={styles.container}>
      <h1 style={styles.heading}>Failed Investment</h1>
      <p style={styles.paragraph}>Sorry, your investment failed. Please try again later.</p>
      <button style={styles.button} onClick={handleRetry}>Retry</button>
      <button style={styles.button} onClick={handleContinue}>Continue</button>
    </div>
  );
};

export default FailurePage;

const styles = {
  container: {
    textAlign: 'center',
    marginTop: '50px',
    fontFamily: 'Arial, sans-serif',
  },
  heading: {
    fontSize: '24px',
    marginBottom: '20px',
    color: '#333',
  },
  paragraph: {
    fontSize: '18px',
    marginBottom: '20px',
  },
  button: {
    backgroundColor: '#007bff',
    color: '#fff',
    border: 'none',
    borderRadius: '5px',
    padding: '10px 20px',
    fontSize: '16px',
    cursor: 'pointer',
    marginRight: '10px',
  },
};
