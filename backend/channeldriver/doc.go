// Package channeldriver defines payment channel (PSP) integration contracts.
//
// Core types and interfaces live in subpackage [base]; this root package re-exports them so
// services can import a single path. One driver implementation is registered per driver_key
// (protocol family). Each row is represented by [BindInput] (channel id, driver key, and the
// raw channel_config JSON string). Platform code may build that JSON using
// [github.com/gloopai/pay/common/channelconfig] (legacy column merge, validation).
//
// [Registry] and [Registry.GetChannelDriver] with a [ChannelResolver] are intended to be wired
// only in the core service. Gateway and trade should call core over gRPC for channel operations;
// they may still import this package for types and HTTP helpers during migration.
//
// Outbound calls use [ChannelDriver] / [BalanceChannel] without passing a
// per-request config blob. Inbound async notifications are verified on the same bound channel.
//
// See README.md in this module for how to add a new channel driver.
package channeldriver
