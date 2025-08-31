// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import "./ConfidentialPayments.sol";
import "./RelayValidator.sol";
import "./TrancheVault.sol";

contract BatchOperations {
    ConfidentialPayments public immutable confidentialPayments;
    RelayValidator public immutable relayValidator;
    TrancheVault public immutable trancheVault;

    struct BatchPayment {
        address recipient;
        address token;
        bytes encryptedAmount;
        string metadataURI;
        bool makePrivate;
    }

    struct BatchValidation {
        uint256 paymentId;
        bytes32 messageHash;
        uint256 amount;
    }

    struct BatchDeposit {
        TrancheVault.TrancheType tranche;
        uint256 amount;
    }

    event BatchPaymentsCreated(uint256[] paymentIds, address indexed sender);
    event BatchValidationsRequested(uint256[] requestIds, address indexed requester);
    event BatchDepositsCompleted(address indexed user, uint256 totalAmount);

    error BatchSizeMismatch();
    error BatchTooLarge();
    error BatchExecutionFailed();

    constructor(
        address _confidentialPayments,
        address _relayValidator,
        address _trancheVault
    ) {
        confidentialPayments = ConfidentialPayments(_confidentialPayments);
        relayValidator = RelayValidator(_relayValidator);
        trancheVault = TrancheVault(_trancheVault);
    }

    function batchCreatePayments(
        BatchPayment[] calldata payments
    ) external payable returns (uint256[] memory paymentIds) {
        if (payments.length == 0 || payments.length > 50) {
            revert BatchTooLarge();
        }

        paymentIds = new uint256[](payments.length);
        uint256 totalValue = 0;

        for (uint256 i = 0; i < payments.length; i++) {
            BatchPayment memory payment = payments[i];
            
            // Calculate required ETH for this payment
            if (payment.token == address(0)) {
                // For ETH payments, we need to estimate the value
                // This is simplified - in practice you'd decode the encrypted amount
                totalValue += 0.1 ether; // Placeholder
            }

            try confidentialPayments.createConfidentialPayment{
                value: payment.token == address(0) ? 0.1 ether : 0
            }(
                payment.recipient,
                payment.token,
                payment.encryptedAmount,
                payment.metadataURI,
                payment.makePrivate
            ) returns (uint256 paymentId) {
                paymentIds[i] = paymentId;
            } catch {
                revert BatchExecutionFailed();
            }
        }

        emit BatchPaymentsCreated(paymentIds, msg.sender);
    }

    function batchRequestValidations(
        BatchValidation[] calldata validations
    ) external returns (uint256[] memory requestIds) {
        if (validations.length == 0 || validations.length > 20) {
            revert BatchTooLarge();
        }

        // Only owner of RelayValidator can request validations
        require(msg.sender == relayValidator.owner(), "Only validator owner");

        requestIds = new uint256[](validations.length);

        for (uint256 i = 0; i < validations.length; i++) {
            BatchValidation memory validation = validations[i];
            
            try relayValidator.requestValidation(
                validation.paymentId,
                validation.messageHash,
                validation.amount
            ) returns (uint256 requestId) {
                requestIds[i] = requestId;
            } catch {
                revert BatchExecutionFailed();
            }
        }

        emit BatchValidationsRequested(requestIds, msg.sender);
    }

    function batchDeposit(
        BatchDeposit[] calldata deposits
    ) external {
        if (deposits.length == 0 || deposits.length > 10) {
            revert BatchTooLarge();
        }

        uint256 totalAmount = 0;

        for (uint256 i = 0; i < deposits.length; i++) {
            BatchDeposit memory deposit = deposits[i];
            totalAmount += deposit.amount;

            try trancheVault.deposit(deposit.tranche, deposit.amount) {
                // Success
            } catch {
                revert BatchExecutionFailed();
            }
        }

        emit BatchDepositsCompleted(msg.sender, totalAmount);
    }

    function batchGrantDisclosurePermissions(
        uint256[] calldata paymentIds,
        address[] calldata viewers
    ) external {
        if (paymentIds.length != viewers.length) {
            revert BatchSizeMismatch();
        }
        if (paymentIds.length > 100) {
            revert BatchTooLarge();
        }

        for (uint256 i = 0; i < paymentIds.length; i++) {
            try confidentialPayments.grantDisclosurePermission(viewers[i], paymentIds[i]) {
                // Success
            } catch {
                revert BatchExecutionFailed();
            }
        }
    }

    function batchRevokeDisclosurePermissions(
        uint256[] calldata paymentIds,
        address[] calldata viewers
    ) external {
        if (paymentIds.length != viewers.length) {
            revert BatchSizeMismatch();
        }
        if (paymentIds.length > 100) {
            revert BatchTooLarge();
        }

        for (uint256 i = 0; i < paymentIds.length; i++) {
            try confidentialPayments.revokeDisclosurePermission(viewers[i], paymentIds[i]) {
                // Success
            } catch {
                revert BatchExecutionFailed();
            }
        }
    }

    function batchCompletePayments(uint256[] calldata paymentIds) external {
        if (paymentIds.length > 50) {
            revert BatchTooLarge();
        }

        for (uint256 i = 0; i < paymentIds.length; i++) {
            try confidentialPayments.completeConfidentialPayment(paymentIds[i]) {
                // Success
            } catch {
                revert BatchExecutionFailed();
            }
        }
    }

    function estimateBatchGas(
        BatchPayment[] calldata payments
    ) external view returns (uint256 totalGas) {
        // Rough estimation: 180k gas per payment + 21k base + batch overhead
        totalGas = 21000 + (payments.length * 185000) + 5000;
    }
}