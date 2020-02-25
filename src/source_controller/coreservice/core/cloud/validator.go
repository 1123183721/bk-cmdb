/*
 * Tencent is pleased to support the open source community by making 蓝鲸 available.,
 * Copyright (C) 2017-2018 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the ",License",); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an ",AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package cloud

import (
	"configcenter/src/common"
	"configcenter/src/common/blog"
	"configcenter/src/common/errors"
	"configcenter/src/common/http/rest"
	"configcenter/src/common/mapstr"
	"configcenter/src/common/metadata"
	"configcenter/src/common/universalsql/mongo"
	"configcenter/src/common/util"
)

func (c *cloudOperation) validCreateAccount(kit *rest.Kit, account *metadata.CloudAccount) errors.CCErrorCoder {
	// accountType check
	if !util.InStrArr(metadata.SupportCloudVendors, string(account.CloudVendor)) {
		blog.ErrorJSON("[validCreateAccount] not support cloud vendor: %s, rid: %v", account.CloudVendor, kit.Rid)
		return kit.CCError.CCErrorf(common.CCErrCloudVendorNotSupport)
	}

	// account name unique check
	cond := mongo.NewCondition()
	cond.Element(&mongo.Eq{Key: common.BKCloudAccountName, Val: account.AccountName})
	count, err := c.countAccount(kit, cond.ToMapStr())
	if nil != err {
		blog.ErrorJSON("[validCreateAccount] count account error %v, condition: %#v, rid: %s", err, cond.ToMapStr(), kit.Rid)
		return kit.CCError.CCError(common.CCErrCommDBSelectFailed)
	}
	if count > 0 {
		blog.ErrorJSON("[validCreateAccount] account name already exist, bk_account_name: %s", account.AccountName, kit.Rid)
		return kit.CCError.CCError(common.CCErrCloudAccountNameAlreadyExist)
	}

	return nil
}

func (c *cloudOperation) validUpdateAccount(kit *rest.Kit, accountID int64, option mapstr.MapStr) errors.CCErrorCoder {
	// accountID exist check
	if err := c.validAccountExist(kit, accountID); err != nil {
		return err
	}
	// accountType check
	if option.Exists(common.BKCloudVendor) {
		cloudVendor, err := option.String(common.BKCloudVendor)
		if err != nil {
			blog.ErrorJSON("[validUpdateAccount] not invalid cloud vendor, option: %v, rid: %v", option, kit.Rid)
			return kit.CCError.CCErrorf(common.CCErrCloudValidAccountParamFail, common.BKCloudVendor)
		}
		if !util.InStrArr(metadata.SupportCloudVendors, cloudVendor) {
			blog.ErrorJSON("[validUpdateAccount] not support cloud vendor: %s, rid: %v", cloudVendor, kit.Rid)
			return kit.CCError.CCError(common.CCErrCloudVendorNotSupport)
		}
	}

	// account name unique check
	if option.Exists(common.BKCloudAccountName) {
		cloudAccountName, err := option.String(common.BKCloudAccountName)
		if err != nil {
			blog.ErrorJSON("[validUpdateAccount] not invalid cloud vendor, option: %v, rid: %v", option, kit.Rid)
			return kit.CCError.CCErrorf(common.CCErrCloudValidAccountParamFail, cloudAccountName)
		}
		cond := mongo.NewCondition()
		cond.Element(&mongo.Eq{Key: common.BKCloudAccountName, Val: cloudAccountName})
		cond.Element(&mongo.Neq{Key: common.BKCloudAccountIDField, Val: accountID})
		count, err := c.countAccount(kit, cond.ToMapStr())
		if nil != err {
			blog.ErrorJSON("[validUpdateAccount] count account error %v, condition: %#v, rid: %s", err, cond.ToMapStr(), kit.Rid)
			return kit.CCError.CCError(common.CCErrCommDBSelectFailed)
		}
		if count > 0 {
			blog.ErrorJSON("[validUpdateAccount] account name already exist, bk_account_name: %s", cloudAccountName, kit.Rid)
			return kit.CCError.CCError(common.CCErrCloudAccountNameAlreadyExist)
		}
	}

	return nil
}

func (c *cloudOperation) validDeleteAccount(kit *rest.Kit, accountID int64) errors.CCErrorCoder {
	// accountID exist check
	if err := c.validAccountExist(kit, accountID); err != nil {
		return err
	}
	return nil
}

func (c *cloudOperation) validAccountExist(kit *rest.Kit, accountID int64) errors.CCErrorCoder {
	cond := mongo.NewCondition()
	cond.Element(&mongo.Eq{Key: common.BKCloudAccountIDField, Val: accountID})
	count, err := c.countAccount(kit, cond.ToMapStr())
	if nil != err {
		blog.ErrorJSON("[validAccountExist] count account error %v, condition: %#v, rid: %s", err, cond.ToMapStr(), kit.Rid)
		return kit.CCError.CCError(common.CCErrCommDBSelectFailed)
	}
	if count <= 0 {
		blog.ErrorJSON("[validAccountExist] no account exist, bk_account_id: %s", accountID, kit.Rid)
		return kit.CCError.CCErrorf(common.CCErrCloudAccountIDNoExistFail, accountID)
	}

	return nil
}
