import { useSearchParams, useNavigate } from 'react-router-dom';
import { useState } from 'react';

const SuccessfulPage = () => {
  const [searchParams] = useSearchParams();
  const transactionId = searchParams.get('paymentId');
  const selectedStrategy = searchParams.get('selectedStrategy');
  const amount = searchParams.get('amount');
  const [isButtonDisabled, setIsButtonDisabled] = useState(false);
  const navigate = useNavigate();
  const phoneNumber = localStorage.getItem('phoneNumber');

  const handleContinueClick = async () => {
    try {
      setIsButtonDisabled(true);
      const response = await fetch('http://localhost:8081/execute-strategy-orders', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          strategyName: selectedStrategy,
          amount: parseFloat(amount),
          paymentID: transactionId,
          phoneNumber: phoneNumber, // Replace with the actual phone number
        }),
      });

      if (response.ok) {
        console.log('Strategy orders executed successfully');
        alert('Strategy orders executed successfully. You will be redirected to the home page in 5 seconds.');

        // Redirect to the home page after 5 seconds
        setTimeout(() => {
          navigate('/home');
        }, 5000);
      } else {
        console.error('Failed to execute strategy orders');
        alert('Failed to execute strategy orders. Please try again.');
      }
    } catch (error) {
      console.error('Error executing strategy orders:', error);
      alert('An error occurred while executing the strategy orders. Please try again later.');
    } finally {
      // Re-enable the button after 5 seconds
      setTimeout(() => {
        setIsButtonDisabled(false);
      }, 5000);
    }
  };

  return (
    <div style={styles.container}>
      <h1 style={styles.heading}>Successful Investment</h1>
      <p style={styles.paragraph}>Congratulations! Your investment was successful.</p>
      {/* <p style={styles.paragraph}>Transaction ID: {transactionId}</p>
      <p style={styles.paragraph}>Selected Strategy: {selectedStrategy}</p> */}
      <button
        style={styles.button}
        onClick={handleContinueClick}
        disabled={isButtonDisabled}
      >
        Continue
      </button>
    </div>
  );
};

export default SuccessfulPage;

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
  },
};