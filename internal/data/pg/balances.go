package pg

import (
	"database/sql"
	"faucet-svc/internal/data"
	"faucet-svc/internal/types/pg"
	"fmt"
	sq "github.com/Masterminds/squirrel"
	"github.com/lib/pq"
	"github.com/pkg/errors"
	"gitlab.com/distributed_lab/kit/pgdb"
)

const (
	balancesTableName  = "balances"
	balancesConstraint = "uniq_idx_balances"
)

var (
	ErrUsersConflict = errors.New("balances primary key conflict")
)

func NewBalancesQ(db *pgdb.DB) data.BalancesQ {
	return &BalancesQ{
		db:  db.Clone(),
		sql: sq.Select("b.*").From(fmt.Sprintf("%s as b", balancesTableName)),
	}
}

type BalancesQ struct {
	db  *pgdb.DB
	sql sq.SelectBuilder
}

func (q *BalancesQ) New() data.BalancesQ {
	return NewBalancesQ(q.db)
}

func (q *BalancesQ) Create(balance *pg.Balance) error {
	stmt := sq.Insert(balancesTableName).SetMap(map[string]interface{}{
		"user_id":       balance.UserId,
		"chain_id":      balance.ChainId,
		"chain_type":    balance.ChainType,
		"token_address": balance.TokenAddress,
		"amount":        balance.Amount,
	})

	err := q.db.Exec(stmt)
	if err != nil {
		pqerr, ok := errors.Cause(err).(*pq.Error)
		if ok {
			if pqerr.Constraint == balancesConstraint {
				return ErrUsersConflict
			}
		}
	}
	return err
}

func (q *BalancesQ) Get() (*pg.Balance, error) {
	var result pg.Balance
	err := q.db.Get(&result, q.sql)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &result, nil
}

func (q *BalancesQ) FilterByUserID(userId string) data.BalancesQ {
	q.sql = q.sql.Where(sq.Eq{"b.user_id": userId})
	return q
}

func (q *BalancesQ) FilterByChainID(chainId string) data.BalancesQ {
	q.sql = q.sql.Where(sq.Eq{"b.chain_id": chainId})
	return q
}

func (q *BalancesQ) FilterByChainType(chainType string) data.BalancesQ {
	q.sql = q.sql.Where(sq.Eq{"b.chain_type": chainType})
	return q
}

func (q *BalancesQ) FilterByTokenAddress(tokenAddress string) data.BalancesQ {
	q.sql = q.sql.Where(sq.Eq{"b.token_address": tokenAddress})
	return q
}

func (q *BalancesQ) Update(balance *pg.Balance) error {
	exist, err := q.FilterByUserID(balance.UserId).
		FilterByChainID(balance.ChainId).
		FilterByChainType(balance.ChainType).
		FilterByTokenAddress(balance.TokenAddress).
		Get()
	if err != nil {
		return err
	}

	if exist == nil {
		err = q.Create(balance)
		return err
	}

	stmt := sq.Update(balancesTableName).
		Set("amount", sq.Expr("amount+?", balance.Amount)).
		Where(sq.Eq{"user_id": balance.UserId}).
		Where(sq.Eq{"chain_id": balance.ChainId}).
		Where(sq.Eq{"chain_type": balance.ChainType}).
		Where(sq.Eq{"token_address": balance.TokenAddress})

	err = q.db.Exec(stmt)
	return err
}
