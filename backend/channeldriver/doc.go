// Package channeldriver defines upstream payment channel (PSP) integration contracts.
//
// Core types and interfaces live in subpackage [base]; this root package re-exports them so
// services can import a single path. One implementation is registered per driver_key
// (protocol family); each row in the channels table supplies a ChannelConfig (keys, gateway
// base URL, channel_id).
//
// Outbound calls use PayinUpstream / PayoutUpstream / BalanceUpstream with an explicit
// ChannelConfig. Inbound async notifications are verified per driver; routing by
// driver_key or channel_id is handled at the gateway layer using the Registry.
//
// See README.md in this module for how to add a new channel driver.
package channeldriver
