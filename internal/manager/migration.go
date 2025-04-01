// SPDX-FileCopyrightText: (C) 2024 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package manager

import (
	"context"
	"database/sql"
	"github.com/open-edge-platform/app-orch-catalog/internal/ent/generated/application"
	"github.com/open-edge-platform/app-orch-catalog/internal/ent/generated/artifact"
	"github.com/open-edge-platform/app-orch-catalog/internal/ent/generated/deploymentpackage"
	"github.com/open-edge-platform/app-orch-catalog/internal/ent/generated/registry"
	"github.com/open-edge-platform/app-orch-catalog/internal/northbound"
	// pq is Postgres driver for the database/sql package
	_ "github.com/lib/pq"

	"github.com/open-edge-platform/app-orch-catalog/internal/ent/generated"
)

type migration struct {
	dbClient *generated.Client
	newUUID  string
}

const oldUUID = "default"

func newMigration(client *generated.Client, newUUID string) *migration {
	return &migration{dbClient: client, newUUID: newUUID}
}

func (m *migration) run(ctx context.Context) error {
	log.Infof("Running project migration to %s", m.newUUID)
	tx, err := m.startTransaction(ctx)
	if err != nil {
		return err
	}

	if err = m.migrate(ctx); err != nil {
		m.rollbackTransaction(tx)
		return err
	}

	if err = m.commitTransaction(tx); err != nil {
		return err
	}

	log.Infof("Project migration to %s completed", m.newUUID)
	return nil
}

func (m *migration) migrate(ctx context.Context) error {
	if err := m.migrateRegistries(ctx); err != nil {
		return err
	}

	if err := m.migrateArtifacts(ctx); err != nil {
		return err
	}

	if err := m.migrateApplications(ctx); err != nil {
		return err
	}

	return m.migratePackages(ctx)
}

func (m *migration) migrateRegistries(ctx context.Context) error {
	if northbound.UseSecretService {
		registriesDB, err := m.dbClient.Registry.Query().Where(registry.ProjectUUID(oldUUID)).All(ctx)
		if err != nil {
			return err
		}

		secretService, err := northbound.SecretServiceFactory(ctx)
		if err != nil {
			return err
		}
		defer secretService.Logout(ctx)

		for _, registryDB := range registriesDB {
			err = m.migrateSecret(ctx, registryDB, secretService)
			if err != nil {
				return err
			}
		}
	}

	return m.dbClient.Registry.Update().
		Where(registry.ProjectUUID(oldUUID)).
		SetProjectUUID(m.newUUID).Exec(ctx)
}

func (m *migration) migrateSecret(ctx context.Context, registryDB *generated.Registry, secretService northbound.SecretService) error {
	oldRegistryKey := northbound.MakeSecretPath(registryDB.ProjectUUID, registryDB.Name)
	newRegistryKey := northbound.MakeSecretPath(m.newUUID, registryDB.Name)

	// Fetch the old stored secret
	registrySecretData, err := secretService.ReadSecret(ctx, oldRegistryKey)
	if err != nil {
		log.Warnf("Unable to read old secret %s: %+v", oldRegistryKey, err)
		return nil // For now suppress any errors
	}

	// Write the new secret
	err = secretService.WriteSecret(ctx, newRegistryKey, registrySecretData)
	if err != nil {
		log.Warnf("Unable to write new secret %s: %+v", newRegistryKey, err)
		return nil // For now suppress any errors
	}

	// Delete the old stored secret
	err = secretService.DeleteSecret(ctx, oldRegistryKey)
	if err != nil {
		log.Warnf("Unable to delete old secret %s: %+v", oldRegistryKey, err)
		return nil // For now suppress any errors
	}

	// For now, suppress any errors during secret service migration operations, opting for warning log entries instead.
	return nil
}

func (m *migration) migrateArtifacts(ctx context.Context) error {
	return m.dbClient.Artifact.Update().
		Where(artifact.ProjectUUID(oldUUID)).
		SetProjectUUID(m.newUUID).Exec(ctx)
}

func (m *migration) migrateApplications(ctx context.Context) error {
	return m.dbClient.Application.Update().
		Where(application.ProjectUUID(oldUUID)).
		SetProjectUUID(m.newUUID).Exec(ctx)
}

func (m *migration) migratePackages(ctx context.Context) error {
	return m.dbClient.DeploymentPackage.Update().
		Where(deploymentpackage.ProjectUUID(oldUUID)).
		SetProjectUUID(m.newUUID).Exec(ctx)
}

// Starts a new transaction or returns ready to punt error
func (m *migration) startTransaction(ctx context.Context) (*generated.Tx, error) {
	tx, err := m.dbClient.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead})
	if err != nil {
		return nil, err
	}
	return tx, nil
}

// Commits the specified transaction or returns ready to punt error
func (m *migration) commitTransaction(tx *generated.Tx) error {
	err := tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

// Rolls back the specified transaction, absorbing any error
func (m *migration) rollbackTransaction(tx *generated.Tx) {
	_ = tx.Rollback()
}
