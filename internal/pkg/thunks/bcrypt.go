// Copyright 2020 Nametag Inc.
//
// All information contained herein is the property of Nametag Inc.. The
// intellectual and technical concepts contained herein are proprietary, trade
// secrets, and/or confidential to Nametag, Inc. and may be covered by U.S.
// and Foreign Patents, patents in process, and are protected by trade secret or
// copyright law. Reproduction or distribution, in whole or in part, is
// forbidden except by express written permission of Nametag, Inc.

package thunks

import "golang.org/x/crypto/bcrypt"

// GenerateFromPassword is a thunk for bcrypt.GenerateFromPassword
var GenerateFromPassword = bcrypt.GenerateFromPassword

// CompareHashAndPassword is a thunk for bcrypt.CompareHashAndPassword
var CompareHashAndPassword = bcrypt.CompareHashAndPassword
